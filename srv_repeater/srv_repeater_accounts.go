package srv_repeater

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gazercloud/gazer_repeater/logger"
	"github.com/gazercloud/gazer_repeater/storage"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func isEmailValid(e string) bool {
	if len(e) < 3 && len(e) > 254 {
		return false
	}
	return emailRegex.MatchString(e)
}

func (c *SrvRepeater) Registration(requestText []byte, addr string) ([]byte, error) {
	logger.Println("[SrvRepeater]", "Registration", addr)

	var err error
	type RegistrationRequest struct {
		EMail    string `json:"email"`
		Password string `json:"password"`
		Token    string `json:"token"`
	}
	var req RegistrationRequest
	err = json.Unmarshal(requestText, &req)
	if err != nil {
		logger.Println("[SrvRepeater]", "[error]", "Registration Unmarshal error:", err)
		return nil, err
	}

	if req.EMail == "" {
		logger.Println("[SrvRepeater]", "[error]", "Registration error:", "e-mail address is empty")
		return nil, errors.New("e-mail address is empty")
	}

	logger.Println("[SrvRepeater]", "Registration email:", req.EMail)

	if len(req.EMail) > 64 {
		logger.Println("[SrvRepeater]", "[error]", "Registration error:", "e-mail address is too long", req.EMail)
		return nil, errors.New("e-mail address is too long")
	}

	if !isEmailValid(req.EMail) {
		logger.Println("[SrvRepeater]", "[error]", "Registration error:", "invalid e-mail", req.EMail)
		return nil, errors.New("invalid e-mail")
	}

	if len(req.Password) < 8 {
		logger.Println("[SrvRepeater]", "[error]", "Registration error:", "password is too simple")
		return nil, errors.New("password is too simple")
	}

	if len(req.Password) > 64 {
		logger.Println("[SrvRepeater]", "[error]", "Registration error:", "password is too long")
		return nil, errors.New("password is too long")
	}

	recaptchaScore := float64(-1)
	var recaptchaResultStr string
	recaptchaResultStr, err = HttpPostCallReCaptcha("https://www.google.com/recaptcha/api/siteverify", req.Token)
	if err == nil {
		type ReCaptchaResult struct {
			Success bool    `json:"success"`
			Score   float64 `json:"score"`
			Action  string  `json:"action"`
		}
		var recaptchaResult ReCaptchaResult
		err = json.Unmarshal([]byte(recaptchaResultStr), &recaptchaResult)
		if err == nil {
			if recaptchaResult.Success {
				recaptchaScore = recaptchaResult.Score
			} else {
				logger.Println("[SrvRepeater]", "[error]", "reCaptcha (success) error: ", err)
				return nil, errors.New("reCaptcha !success")
			}
		} else {
			logger.Println("[SrvRepeater]", "[error]", "reCaptcha (parsing) error: ", err)
			return nil, err
		}
	} else {
		logger.Println("[SrvRepeater]", "[error]", "reCaptcha (http request) error: ", err)
		return nil, err
	}

	/*
		{
		  "success": true,
		  "challenge_ts": "2021-06-19T15:50:54Z",
		  "hostname": "home.gazer.cloud",
		  "score": 0.9,
		  "action": "submit"
		}
	*/

	if recaptchaScore < 0.5 {
		logger.Println("[SrvRepeater]", "[error]", "Registration ReCaptcha level error:", recaptchaScore)
		return nil, errors.New("you did not pass the ReCaptcha check. Your scores is " + fmt.Sprint(recaptchaScore))
	}

	hash := sha256.New()
	hash.Write([]byte(req.Password))
	shaPassword := hex.EncodeToString(hash.Sum(nil))

	var confirmKey string
	confirmKey, err = c.storage.Registration(req.EMail, shaPassword, addr, recaptchaResultStr, recaptchaScore)
	if err != nil {
		logger.Println("[SrvRepeater]", "[error]", "Registration storage.Registration error:", err)
		return nil, err
	}

	emailContent := `
<div style="text-align: center; color: #FFFFFF;">
<div style="margin: 0 auto; max-width: 600px;">

<div style="font-family: arial; text-align: center; margin: 0px; padding: 20px; color: #EEEEEE; background-color: #333333; font-size: 24pt; border-bottom: solid #D27607 1px">
	<div style="margin: 0 auto;">
		<table style="margin: 0 auto; border-spacing: 0px;border-collapse: separate; max-width: 64px; max-height: 64px;">
			<tr>
				<td style="padding: 2px;"><div style="background-color: #00A0E3; min-width: 30px; max-width: 30px; min-height: 30px"/></td>
				<td style="padding: 2px;"><div style="background-color: #19BB4F; min-width: 30px; max-width: 30px; min-height: 30px"/></td>
			</tr>
			<tr>
				<td style="padding: 2px;"><div style="background-color: #00A0E3; min-width: 30px; max-width: 30px; min-height: 30px"/></td>
				<td style="padding: 2px;"><div style="background-color: #00A0E3; min-width: 30px; max-width: 30px; min-height: 30px"/></td>
			</tr>
		</table>
	</div>
	<div style="font-size: 16pt; margin-top: 10px;">
		GAZER CLOUD
	</div>
</div>

<div style="background-color: #222222; margin: 0px; padding: 20px;">

<div style="text-align: center; padding: 20px; font-family: arial; font-size: 16pt;">
	<h1 style='color: #008800;'>Please confirm your Gazer.Cloud account</h1>
	<div>Please confirm your Gazer.Cloud account and newsletter subscriptions by verifying your email address.</div>
</div>

<div style="padding: 20px;  font-size: 16pt; text-align: center;">
	<a style="text-decoration: none; color: #FFFFFF; font-family: arial;" href="https://home.gazer.cloud/api/request?fn=s-confirm-registration&amp;rj={&quot;key&quot;:&quot;#KEY#&quot;}"><div style="display: inline-block; margin: 10px;background-color: #2C873A; padding: 20px; min-width: 250px; max-width: 250px; text-align: center;">CONFIRM</div></a>
</div>

<div style="text-align: center; padding: 20px; font-family: arial; font-size: 16pt;">
	For all questions you can contact us by e-mail: admin@gazer.cloud
</div>

<div style="text-align: center; padding: 20px; font-family: arial; font-size: 16pt;">
	Best regards, GazerCloud Team.
</div>
</div>

</div>
</div>
<div>URL for confirmation: https://home.gazer.cloud/api/request?fn=s-confirm-registration&amp;rj={&quot;key&quot;:&quot;#KEY#&quot;}</div>
`
	emailContent = strings.ReplaceAll(emailContent, "#KEY#", confirmKey)

	err = SendEMail("Gazer.Cloud - confirm registration", req.EMail, emailContent)
	if err != nil {
		logger.Println("[SrvRepeater]", "[error]", "Registration SendEMail error:", err)
		return nil, err
	}

	type RegistrationResponse struct {
	}
	var bs []byte
	var resp RegistrationResponse
	bs, err = json.MarshalIndent(resp, "", " ")
	if err != nil {
		logger.Println("[SrvRepeater]", "[error]", "Registration json.MarshalIndent error:", err)
		return nil, err
	}

	c.storage.Log("i", "reg1 ok email: "+req.EMail)

	return bs, nil
}

func (c *SrvRepeater) ConfirmRegistration(requestText []byte, addr string) ([]byte, error) {
	logger.Println("[SrvRepeater]", "ConfirmRegistration")

	var err error
	type ConformRegistrationRequest struct {
		Key string `json:"key"`
	}
	var req ConformRegistrationRequest
	err = json.Unmarshal(requestText, &req)
	if err != nil {
		logger.Println("[SrvRepeater]", "[error]", "ConfirmRegistration json.Unmarshal error:", err)
		return nil, err
	}

	if len(req.Key) < 1 || len(req.Key) > 512 {
		logger.Println("[SrvRepeater]", "[error]", "ConfirmRegistration wrong confirmation key")
		return nil, errors.New("wrong confirmation key")
	}

	logger.Println("[SrvRepeater]", "ConfirmRegistration key:", req.Key)

	var email string
	email, err = c.storage.ConfirmRegistration(req.Key)
	if err != nil {
		logger.Println("[SrvRepeater]", "[error]", "ConfirmRegistration ConfirmRegistration error:", err)
		c.storage.Log("e", "reg2 err email: "+email+" key: "+req.Key)
		return nil, err
	}

	type ConformRegistrationResponse struct {
	}
	var bs []byte
	var resp ConformRegistrationResponse
	bs, err = json.MarshalIndent(resp, "", " ")
	if err != nil {
		logger.Println("[SrvRepeater]", "[error]", "ConfirmRegistration json.MarshalIndent error:", err)
		return nil, err
	}

	c.storage.Log("i", "reg2 ok email: "+email)
	return bs, nil
}

func (c *SrvRepeater) ChangePassword(session *storage.Session, requestText []byte, addr string) ([]byte, error) {
	logger.Println("[SrvRepeater]", "ChangePassword")

	var err error
	type ChangePasswordRequest struct {
		Password string `json:"password"`
	}
	var req ChangePasswordRequest
	err = json.Unmarshal(requestText, &req)
	if err != nil {
		return nil, err
	}

	if len(req.Password) < 8 {
		logger.Println("[SrvRepeater]", "[error]", "ChangePassword error:", "password is too simple")
		return nil, errors.New("password is too simple")
	}

	if len(req.Password) > 64 {
		logger.Println("[SrvRepeater]", "[error]", "ChangePassword error:", "password is too long")
		return nil, errors.New("password is too long")
	}

	hash := sha256.New()
	hash.Write([]byte(req.Password))
	shaPassword := hex.EncodeToString(hash.Sum(nil))

	err = c.storage.ChangePassword(session, shaPassword)
	if err != nil {
		logger.Println("[SrvRepeater]", "[error]", "ChangePassword storage.ChangePassword error:", err)
		return nil, err
	}

	type ChangePasswordResponse struct {
	}
	var bs []byte
	var resp ChangePasswordResponse
	bs, err = json.MarshalIndent(resp, "", " ")
	if err != nil {
		return nil, err
	}

	c.storage.Log("i", "change password ok user_id: "+fmt.Sprint(session.UserId))

	return bs, nil
}

func (c *SrvRepeater) ResetPassword(session *storage.Session, requestText []byte, addr string) ([]byte, error) {
	logger.Println("[SrvRepeater]", "ResetPassword")

	var err error
	type ResetPasswordRequest struct {
		Token    string `json:"token"`
		Key      string `json:"key"`
		Password string `json:"password"`
	}
	var req ResetPasswordRequest
	err = json.Unmarshal(requestText, &req)
	if err != nil {
		logger.Println("[SrvRepeater]", "[error]", "ResetPassword json.Unmarshal error:", err)
		return nil, err
	}

	logger.Println("[SrvRepeater]", "ResetPassword key:", req.Key)

	if len(req.Password) < 8 {
		logger.Println("[SrvRepeater]", "[error]", "ResetPassword error:", "password is too simple")
		return nil, errors.New("password is too simple")
	}

	if len(req.Password) > 64 {
		logger.Println("[SrvRepeater]", "[error]", "ResetPassword error:", "password is too long")
		return nil, errors.New("password is too long")
	}

	recaptchaScore := float64(-1)
	var recaptchaResultStr string
	recaptchaResultStr, err = HttpPostCallReCaptcha("https://www.google.com/recaptcha/api/siteverify", req.Token)
	if err == nil {
		type ReCaptchaResult struct {
			Success bool    `json:"success"`
			Score   float64 `json:"score"`
			Action  string  `json:"action"`
		}
		var recaptchaResult ReCaptchaResult
		err = json.Unmarshal([]byte(recaptchaResultStr), &recaptchaResult)
		if err == nil {
			if recaptchaResult.Success {
				recaptchaScore = recaptchaResult.Score
			} else {
				logger.Println("[SrvRepeater]", "[error]", "ResetPassword reCaptcha (success) error: ", err)
				return nil, errors.New("reCaptcha !success")
			}
		} else {
			logger.Println("[SrvRepeater]", "[error]", "ResetPassword reCaptcha (parsing) error: ", err)
			return nil, err
		}
	} else {
		logger.Println("[SrvRepeater]", "[error]", "ResetPassword reCaptcha (http request) error: ", err)
		return nil, err
	}

	if recaptchaScore < 0.5 {
		logger.Println("[SrvRepeater]", "[error]", "ResetPassword reCaptcha (http request) level error: ", recaptchaScore)
		return nil, errors.New("you did not pass the ReCaptcha check. Your scores is " + fmt.Sprint(recaptchaScore))
	}

	hash := sha256.New()
	hash.Write([]byte(req.Password))
	shaPassword := hex.EncodeToString(hash.Sum(nil))

	err = c.storage.ChangePasswordByKey(req.Key, shaPassword)
	if err != nil {
		logger.Println("[SrvRepeater]", "[error]", "ResetPassword storage.ChangePasswordByKey error:", err)
		return nil, err
	}

	type ResetPasswordResponse struct {
	}
	var bs []byte
	var resp ResetPasswordResponse
	bs, err = json.MarshalIndent(resp, "", " ")
	if err != nil {
		logger.Println("[SrvRepeater]", "[error]", "ResetPassword json.MarshalIndent error:", err)
		return nil, err
	}

	c.storage.Log("i", "reset password 2 ok user_id: "+fmt.Sprint(session.UserId))

	return bs, nil
}

func (c *SrvRepeater) RestorePassword(requestText []byte, addr string) ([]byte, error) {
	logger.Println("[SrvRepeater]", "RestorePassword")

	var err error
	type RestoreRequest struct {
		Email string `json:"email"`
		Token string `json:"token"`
	}
	var req RestoreRequest
	err = json.Unmarshal(requestText, &req)
	if err != nil {
		logger.Println("[SrvRepeater]", "[error]", "RestorePassword json.Unmarshal error: ", err)
		return nil, err
	}

	if req.Email == "" {
		logger.Println("[SrvRepeater]", "[error]", "RestorePassword error: ", "e-mail address is empty")
		return nil, errors.New("e-mail address is empty")
	}

	logger.Println("[SrvRepeater]", "RestorePassword email:", req.Email)

	if len(req.Email) > 64 {
		logger.Println("[SrvRepeater]", "[error]", "RestorePassword error: ", "e-mail address is too long", req.Email)
		return nil, errors.New("e-mail address is too long")
	}

	if !isEmailValid(req.Email) {
		logger.Println("[SrvRepeater]", "[error]", "RestorePassword error: ", "invalid e-mail", req.Email)
		return nil, errors.New("invalid e-mail")
	}

	recaptchaScore := float64(-1)
	var recaptchaResultStr string
	recaptchaResultStr, err = HttpPostCallReCaptcha("https://www.google.com/recaptcha/api/siteverify", req.Token)
	if err == nil {
		type ReCaptchaResult struct {
			Success bool    `json:"success"`
			Score   float64 `json:"score"`
			Action  string  `json:"action"`
		}
		var recaptchaResult ReCaptchaResult
		err = json.Unmarshal([]byte(recaptchaResultStr), &recaptchaResult)
		if err == nil {
			if recaptchaResult.Success {
				recaptchaScore = recaptchaResult.Score
			} else {
				logger.Println("[SrvRepeater]", "[error]", "RestorePassword reCaptcha (success) error: ", err)
				return nil, errors.New("reCaptcha !success")
			}
		} else {
			logger.Println("[SrvRepeater]", "[error]", "RestorePassword reCaptcha (parsing) error: ", err)
			return nil, err
		}
	} else {
		logger.Println("[SrvRepeater]", "[error]", "RestorePassword reCaptcha (http request) error: ", err)
		return nil, err
	}

	if recaptchaScore < 0.5 {
		logger.Println("[SrvRepeater]", "[error]", "RestorePassword ReCaptcha level error:", recaptchaScore)
		return nil, errors.New("you did not pass the ReCaptcha check. Your scores is " + fmt.Sprint(recaptchaScore))
	}

	var currentPassword string
	currentPassword, err = c.storage.GetPassword(req.Email)
	if err != nil {
		logger.Println("[SrvRepeater]", "[error]", "RestorePassword storage.GetPassword error:", err)
		return nil, err
	}

	type ResetPasswordStruct struct {
		EMail    string `json:"e_mail"`
		Password string `json:"password"`
	}

	var keyObject ResetPasswordStruct
	keyObject.EMail = req.Email
	keyObject.Password = currentPassword
	keyBS, _ := json.Marshal(keyObject)
	key := hex.EncodeToString(keyBS)

	emailContent := `

<div style="text-align: center; color: #FFFFFF;">
<div style="margin: 0 auto; max-width: 600px;">

<div style="font-family: arial; text-align: center; margin: 0px; padding: 20px; color: #EEEEEE; background-color: #333333; font-size: 24pt; border-bottom: solid #D27607 1px">
	<div style="margin: 0 auto;">
		<table style="margin: 0 auto; border-spacing: 0px;border-collapse: separate; max-width: 64px; max-height: 64px;">
			<tr>
				<td style="padding: 2px;"><div style="background-color: #00A0E3; min-width: 30px; max-width: 30px; min-height: 30px"/></td>
				<td style="padding: 2px;"><div style="background-color: #19BB4F; min-width: 30px; max-width: 30px; min-height: 30px"/></td>
			</tr>
			<tr>
				<td style="padding: 2px;"><div style="background-color: #00A0E3; min-width: 30px; max-width: 30px; min-height: 30px"/></td>
				<td style="padding: 2px;"><div style="background-color: #00A0E3; min-width: 30px; max-width: 30px; min-height: 30px"/></td>
			</tr>
		</table>
	</div>
	<div style="font-size: 16pt; margin-top: 10px;">
		GAZER CLOUD
	</div>
</div>

<div style="background-color: #222222; margin: 0px; padding: 20px;">

<div style="text-align: center; padding: 20px; font-family: arial; font-size: 16pt;">
	<div>To reset your password, please click the link below.</div>
</div>

<div style="padding: 20px;  font-size: 16pt; text-align: center;">
	<a style="text-decoration: none; color: #FFFFFF; font-family: arial;" href="https://home.gazer.cloud/#form=reset_password&amp;key=#KEY#"><div style="display: inline-block; margin: 10px;background-color: #2C873A; padding: 20px; min-width: 250px; max-width: 250px; text-align: center;">RESET PASSWORD</div></a>
</div>

<div style="text-align: center; padding: 20px; font-family: arial; font-size: 16pt;">
	For all questions you can contact us by e-mail: admin@gazer.cloud
</div>

<div style="text-align: center; padding: 20px; font-family: arial; font-size: 16pt;">
	Best regards, GazerCloud Team.
</div>
</div>

</div>
</div>
`
	emailContent = strings.ReplaceAll(emailContent, "#KEY#", key)

	err = SendEMail("Gazer.Cloud - Password Reset Instructions", req.Email, emailContent)
	if err != nil {
		logger.Println("[SrvRepeater]", "[error]", "RestorePassword SendEMail error:", err)
		return nil, err
	}

	type RestorePasswordResponse struct {
	}
	var bs []byte
	var resp RestorePasswordResponse
	bs, err = json.MarshalIndent(resp, "", " ")
	if err != nil {
		logger.Println("[SrvRepeater]", "[error]", "RestorePassword json.MarshalIndent error:", err)
		return nil, err
	}

	c.storage.Log("i", "reset password 1 ok email: "+req.Email)

	return bs, nil
}

func (c *SrvRepeater) NodeAdd(session *storage.Session, requestText []byte, addr string) (bs []byte, err error) {
	logger.Println("[SrvRepeater]", "NodeAdd")

	if session == nil {
		err = errors.New("no session")
		return
	}

	type NodeAddRequest struct {
		Name string `json:"name"`
	}
	var req NodeAddRequest
	err = json.Unmarshal(requestText, &req)
	if err != nil {
		logger.Println("[SrvRepeater]", "[error]", "NodeAdd name:", req.Name, err)
		return
	}

	logger.Println("[SrvRepeater]", "NodeAdd name:", req.Name)

	var nodeId string
	nodeId, err = c.storage.NodeAdd(session.UserId, req.Name)
	if err != nil {
		logger.Println("[SrvRepeater]", "[error]", "NodeAdd userId", session.UserId, "name:", req.Name, err)
		return
	}

	type NodeAddResponse struct {
		Id string `json:"id"`
	}
	var resp NodeAddResponse
	resp.Id = nodeId
	bs, err = json.Marshal(resp)

	c.storage.Log("i", "add node user_id: "+fmt.Sprint(session.UserId))

	return
}

func (c *SrvRepeater) NodeUpdate(session *storage.Session, requestText []byte, addr string) (bs []byte, err error) {
	logger.Println("[SrvRepeater]", "NodeUpdate")

	if session == nil {
		err = errors.New("no session")
		return
	}

	type NodeUpdateRequest struct {
		NodeId string `json:"node_id"`
		Name   string `json:"name"`
	}
	var req NodeUpdateRequest
	err = json.Unmarshal(requestText, &req)
	if err != nil {
		logger.Println("[SrvRepeater]", "[error]", "NodeUpdate json.Unmarshal error", err)
		return
	}

	err = c.storage.NodeUpdate(session.UserId, req.NodeId, req.Name)
	if err != nil {
		logger.Println("[SrvRepeater]", "[error]", "NodeUpdate storage.NodeUpdate error", err)
		return
	}

	type NodeUpdateResponse struct {
	}
	var resp NodeUpdateResponse
	bs, err = json.Marshal(resp)

	c.storage.Log("i", "update node user_id: "+fmt.Sprint(session.UserId))

	return
}

func (c *SrvRepeater) NodeRemove(session *storage.Session, requestText []byte, addr string) (bs []byte, err error) {
	logger.Println("[SrvRepeater]", "NodeRemove")

	if session == nil {
		err = errors.New("no session")
		return
	}

	type NodeUpdateRequest struct {
		NodeId string `json:"node_id"`
	}
	var req NodeUpdateRequest
	err = json.Unmarshal(requestText, &req)
	if err != nil {
		logger.Println("[SrvRepeater]", "[error]", "NodeRemove json.Unmarshal error", err)
		return
	}

	err = c.storage.NodeRemove(session.UserId, req.NodeId)
	if err != nil {
		logger.Println("[SrvRepeater]", "[error]", "NodeRemove storage.NodeRemove error", err)
		return
	}

	type NodeUpdateResponse struct {
	}
	var resp NodeUpdateResponse
	bs, err = json.Marshal(resp)

	c.storage.Log("i", "remove node user_id: "+fmt.Sprint(session.UserId))

	return
}
