package mail

import (
	"testing"

	"github.com/rdozzi/simple_bank/db/util"
	"github.com/stretchr/testify/require"
)

func TestSendEmailWithGmail(t *testing.T){
	if testing.Short(){
		t.Skip()
	}
	config, err := util.LoadConfig("..")
	require.NoError(t,err)

	sender := NewGmailSender(config.EmailSenderName,config.EmailSenderAddress,config.EmailSenderPassword)

	subject := "A test email"
	content := `
	<h1>Hello world</h1>
	<p>This is a test message from the Simple Bank Application</p>
	`

	to := []string{"rdozzi84@gmail.com"}
	// attachFiles := []string("../README.md")

	err = sender.SendEmail(subject,content,to,nil,nil,nil)
	require.NoError(t,err)

}