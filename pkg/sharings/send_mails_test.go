package sharings

import (
	"fmt"
	"testing"

	"github.com/cozy/cozy-stack/pkg/consts"
	"github.com/cozy/cozy-stack/pkg/couchdb"
	"github.com/cozy/cozy-stack/pkg/oauth"
	"github.com/cozy/cozy-stack/pkg/permissions"
	"github.com/cozy/cozy-stack/web/jsonapi"
	"github.com/stretchr/testify/assert"
)

var rec = &Recipient{
	URL: "",
	Client: &oauth.Client{
		ClientID:     "",
		RedirectURIs: []string{},
	},
}

var recStatus = &RecipientStatus{
	RefRecipient: jsonapi.ResourceIdentifier{
		Type: consts.Recipients,
	},
}

var mailValues = &mailTemplateValues{}

var sharing = &Sharing{
	SharingType:      consts.OneShotSharing,
	RecipientsStatus: []*RecipientStatus{recStatus},
	SharingID:        "sparta-id",
	Permissions:      &permissions.Set{},
}

func TestLogErrorAndSetRecipientStatus(t *testing.T) {
	err := ErrMailCouldNotBeSent
	res := logErrorAndSetRecipientStatus(recStatus, err)

	assert.Equal(t, true, res)
	assert.Equal(t, consts.ErrorStatus, recStatus.Status)
}

func TestGenerateMailMessageWhenRecipientHasNoEmail(t *testing.T) {
	msg, err := generateMailMessage(sharing, rec, mailValues)
	assert.Error(t, err)
	assert.Equal(t, ErrRecipientHasNoEmail, err)
	assert.Nil(t, msg)
}

func TestGenerateMailMessageSuccess(t *testing.T) {
	rec.Email = "this@is.mail"
	_, err := generateMailMessage(sharing, rec, mailValues)
	assert.NoError(t, err)
}

func TestGenerateOAuthQueryStringWhenThereIsNoOAuthClient(t *testing.T) {
	// Without client id.
	oauthQueryString, err := generateOAuthQueryString(sharing, rec)
	assert.Error(t, err)
	assert.Equal(t, ErrNoOAuthClient, err)
	assert.Equal(t, oauthQueryString, "")

	// Without redirect uri.
	rec.Client.ClientID = "sparta"
	rec.Client.RedirectURIs = []string{}
	oauthQueryString, err = generateOAuthQueryString(sharing, rec)
	assert.Error(t, err)
	assert.Equal(t, ErrNoOAuthClient, err)
	assert.Equal(t, oauthQueryString, "")

}

func TestGenerateOAuthQueryStringWhenRecipientHasNoURL(t *testing.T) {
	rec.Client.RedirectURIs = []string{"redirect.me.to.sparta"}

	oauthQueryString, err := generateOAuthQueryString(sharing, rec)
	assert.Error(t, err)
	assert.Equal(t, ErrRecipientHasNoURL, err)
	assert.Equal(t, "", oauthQueryString)
}

func TestGenerateOAuthQueryStringSuccess(t *testing.T) {
	rec.URL = "this.is.url"

	_, err := generateOAuthQueryString(sharing, rec)
	assert.NoError(t, err)
}

func TestSendSharingMails(t *testing.T) {
	// We provoke the error that occurrs when a recipient has no URL or no
	// OAuth client by creating an incomplete recipient document.
	rec.URL = ""
	// Add the recipient in the database.
	err := couchdb.CreateDoc(in, rec)
	if err != nil {
		fmt.Printf("%v\n", err)
		t.Fail()
	}
	defer couchdb.DeleteDoc(in, rec)
	// Set the id to the id generated by Couch.
	recStatus.RefRecipient.ID = rec.RID

	err = SendSharingMails(in, sharing)
	assert.Error(t, err)
	assert.Equal(t, ErrMailCouldNotBeSent, err)

	// The other scenario is when the recipient has no email set.
	rec.URL = "this.is.url"
	rec.Email = ""
	err = couchdb.UpdateDoc(in, rec)
	if err != nil {
		fmt.Printf("%v\n", err)
		t.Fail()
	}

	err = SendSharingMails(in, sharing)
	assert.Error(t, err)
	assert.Equal(t, ErrMailCouldNotBeSent, err)
}
