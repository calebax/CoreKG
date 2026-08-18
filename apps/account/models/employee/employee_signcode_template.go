package employee

const emailBodyTemplate = `<div>
  <table
    style="
      color: rgb(0, 0, 0);
      font-family: 'helvetica neue', PingFangSC-Light, arial, 'hiragino sans gb',
        'microsoft yahei ui', 'microsoft yahei', simsun, sans-serif;
      font-size: 14px;
      text-align: center;
      background-color: rgb(247, 248, 250);
      width: 658px;
      margin-bottom: 10px;
      border-collapse: collapse;
    "
  >
    <tbody>
      <tr>
        <td
          style="
            font-family: 'lucida Grande', Verdana;
            font-size: 12px;
            -webkit-font-smoothing: subpixel-antialiased;
            width: 17.7344px;
            max-width: 30px;
          "
        ></td>
        <td
          style="
            font-family: 'lucida Grande', Verdana;
            font-size: 12px;
            -webkit-font-smoothing: subpixel-antialiased;
            max-width: 600px;
          "
        >
          <p
            style="
              line-height: 0px;
              height: 2px;
              background-color: rgb(0, 164, 255);
              border: 0px;
              font-size: 0px;
              padding: 0px;
              width: 616.531px;
              margin-top: 20px;
            "
          ></p>
          <div
            id="cTMail-inner"
            style="
              background-color: rgb(255, 255, 255);
              padding: 23px 0px 20px;
              box-shadow: rgba(122, 55, 55, 0.2) 0px 1px 1px 0px;
              text-align: left;
            "
          >
            <table
              style="
                width: 616.531px;
                margin-bottom: 10px;
                border-collapse: collapse;
              "
            >
              <tbody>
                <tr>
                  <td
                    style="
                      font-size: 12px;
                      -webkit-font-smoothing: subpixel-antialiased;
                      width: 17.7188px;
                      max-width: 30px;
                    "
                  >
                    <br />
                  </td>
                  <td
                    style="
                      font-size: 12px;
                      -webkit-font-smoothing: subpixel-antialiased;
                      max-width: 480px;
                    "
                  >
                    <p
                      class="cTMail-content"
                      style="
                        line-height: 24px;
                        font-size: 14px;
                        color: rgb(51, 51, 51);
                        margin: 6px 0px 0px;
                        overflow-wrap: break-word;
                        word-break: break-all;
                      "
                    >
                      验证码：{{ .Code }}，该验证码 {{ .Minutes }} 分钟内有效。为了保障您的账户安全，请勿向他人泄漏验证码信息。
                    </p>
                    <br />
                    <dl
                      style="
                        font-size: 14px;
                        color: rgb(51, 51, 51);
                        line-height: 18px;
                      "
                    ></dl>
                    <p
                      id="cTMail-sender"
                      style="
                        line-height: 26px;
                        color: rgb(51, 51, 51);
                        font-size: 14px;
                        overflow-wrap: break-word;
                        word-break: break-all;
                        margin-top: 32px;
                      "
                    >
                      此致<br /><strong>{{ .From }}</strong>
                    </p>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </td>
      </tr>
    </tbody>
  </table>
</div>
`
