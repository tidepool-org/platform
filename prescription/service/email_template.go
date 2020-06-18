package service

import "github.com/tidepool-org/mailer/mailer"

const prescriptionTemplateName string = "prescription"
const prescriptionEmailSubject string = `Your Tidepool Loop Prescription`
const prescriptionEmailBody string = `
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
  <head>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8">
    <!--[if !mso]><!-->
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <!--<![endif]-->
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title></title>
    <!--[if (gte mso 9)|(IE)]>
      <style type="text/css">
        table {border-collapse: collapse;}
      </style>
    <![endif]-->
    <style type="text/css">
      /* Media Queries */
      @media screen and (max-width: 360px) {
        p.attribution {
          font-size: 10px;
          padding: 0 0 0 4px;
        }
      }
    </style>
  </head>
  <body style="padding:0;background-color:#ffffff;font-family:'Open Sans', 'Helvetica Neue', Helvetica, sans-serif;Margin:8px !important;">
    <center class="wrapper" style="width:100%;table-layout:fixed;-webkit-text-size-adjust:100%;-ms-text-size-adjust:100%;">
      <div class="webkit" style="max-width:560px;margin:0 auto;background-color:#F5F5F5;">
        <!--[if (gte mso 9)|(IE)]>
        <table bgcolor="#F5F5F5" width="560" cellpadding="0" cellspacing="0" border="0" align="center">
        <tr>
        <td>
        <![endif]-->
        <table class="outer" align="center" style="border-spacing:0;color:#333333;Margin:0 auto;width:95%;max-width:560px;padding-top:42px;padding-bottom:15px;">
          <tr>
            <td class="one-column" style="padding:0;">
              <table width="100%" style="border-spacing:0;color:#333333;">
                <tr>
                  <td class="inner centered" style="padding:0;padding:10px;text-align:center;">
                    <p class="h1 content-width" style="color:#281946;font-size:14px;line-height:1.5;Margin:0;Margin-bottom:10px;font-size:18px;font-weight:600;Margin-bottom:32px;Margin-left:auto;Margin-right:auto;max-width:400px;">
                      Hey there!
                    </p>
                    <p class="h2 content-width" style="color:#281946;line-height:1.5;Margin:0;Margin-bottom:10px;font-size:14px;font-weight:600;Margin-bottom:28px;Margin-left:auto;Margin-right:auto;max-width:400px;">
                      You have a new Tidepool Loop prescription.<br/><br/>You can retrieve it in the Tidepool Loop application using the following access code: <strong>{{ .AccessCode }}</strong>.
                    </p>
                  </td>
                </tr>
                <tr>
                  <td class="inner centered" style="padding:0;padding:10px;text-align:center;">
                    <p style="color:#281946;font-size:14px;line-height:1.5;font-weight:600;Margin:0;Margin-bottom:10px;">Sincerely,<br/>The Tidepool Team</p>
                  </td>
                </tr>
                <tr>
                  <td class="inner centered" style="padding:0;padding:10px;text-align:center;">
                    <a href="{{ .WebURL }}" style="color:#627CFF;text-decoration:none;"><img class="logo" width="220" height="24" src="{{ .AssetURL }}/img/tidepool_logo_light_x2.png" alt="Tidepool logo" style="border:0;display:inline-block;Margin-bottom:36px;max-width:220px;height:auto;"/></a>
                  </td>
                </tr>
                <tr>
                  <td class="inner centered" style="padding:0;padding:10px;text-align:center;">
                    <table class="links primary" align="center" style="border-spacing:0;color:#333333;">
                      <tr>
                        <td class="no-left-padding" valign="middle" style="padding:0;padding:0 8px;padding-left:0;">
                          <a href="https://www.twitter.com/Tidepool_org" style="color:#627CFF;text-decoration:none;">
                            <img width="32" height="24" src="{{ .AssetURL }}/img/twitter_white_x2.png" alt="Twitter logo" style="border:0;"/>
                          </a>
                        </td>
                        <td valign="middle" style="padding:0;padding:0 8px;">
                          <a href="http://www.facebook.com/TidepoolOrg" style="color:#627CFF;text-decoration:none;">
                            <img width="14" height="24" src="{{ .AssetURL }}/img/facebook_white_x2.png" alt="Facebook logo" style="border:0;"/>
                          </a>
                        </td>
                        <td class="no-right-padding" valign="middle" style="padding:0;padding:0 8px;padding-right:0;">
                          <p class="attribution" style="color:#281946;line-height:1.5;font-weight:600;Margin:0;Margin-bottom:10px;font-size:14px;color:#9b9b9b;padding:0 0 0 8px;Margin-bottom:0;">Made possible by</p>
                        </td>
                        <td class="no-right-padding" valign="middle" style="padding:0;padding:0 8px;padding-right:0;">
                          <a href="http://www.jdrf.org/" style="color:#627CFF;text-decoration:none;">
                            <img width="94" height="24" src="{{ .AssetURL }}/img/jdrf_logo_reverse_x2.png" alt="JDRF logo" style="border:0;"/>
                          </a>
                        </td>
                      </tr>
                    </table>
                  </td>
                </tr>
                <tr>
                  <td class="inner centered" style="padding:0;padding:10px;text-align:center;">
                    <p class="about content-width narrow" style="color:#281946;font-size:14px;line-height:1.5;font-weight:600;Margin:0;Margin-bottom:10px;font-size:10px;font-weight:300;color:#6d6d6d;Margin-bottom:0;max-width:400px;Margin-left:auto;Margin-right:auto;max-width:350px;">
                      <a href="https://www.tidepool.org" style="color:#627CFF;text-decoration:none;">Tidepool</a>
                      An open source, not-for-profit effort to build an open data platform and better applications that reduce the burden of diabetes.
                    </p>
                  </td>
                </tr>
                <tr>
                  <td class="inner centered" style="padding:0;padding:10px;text-align:center;">
                    <table class="links secondary" align="center" style="border-spacing:0;color:#333333;">
                      <tr>
                        <td height="24" class="no-left-padding" valign="top" style="padding:0;padding:0 2px;padding-left:0;">
                          <!--[if (gte mso 9)|(IE)]>
                          <table bgcolor="#FFFFFF">
                          <tr>
                          <td>
                          <![endif]-->
                          <a class="btn secondary small" href="http://support.tidepool.org" style="color:#627CFF;text-decoration:none;border-radius:4px;font-size:14px;font-weight:bold;padding:10px 20px;display:inline-block;border:1px solid #dbdee0;background-color:#FFFFFF;color:#281946;font-weight:normal;padding:4px 10px 5px;Margin-left:3px;Margin-right:3px;font-size:10px;border-radius:2px;">
                            Get Support
                          </a>
                          <!--[if (gte mso 9)|(IE)]>
                          </td>
                          </tr>
                          </table>
                          <![endif]-->
                        </td>
                      </tr>
                    </table>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
        </table>
        <!--[if (gte mso 9)|(IE)]>
        </td>
        </tr>
        </table>
        <![endif]-->
      </div>
    </center>
  </body>
</html>
`

func NewPrescriptionEmailTemplate() (*mailer.EmailTemplate, error) {
	return mailer.NewEmailTemplate(prescriptionTemplateName, prescriptionEmailSubject, prescriptionEmailBody)
}
