package main

import (
  "fmt"
  "io/ioutil"
  "log"
  "net/http"
  "net/url"
  "strings"

  "golang.org/x/oauth2"
  "golang.org/x/oauth2/facebook"
)

var (
  oauthConf = &oauth2.Config{
    ClientID:     "2297073320579124",
    ClientSecret: "12eaecc719e487841b52892f8d0a89dc",
    RedirectURL:  "http://localhost:9090/oauth2callback",
    Scopes:       []string{"public_profile", "email"},
    Endpoint:     facebook.Endpoint,
  }
  oauthStateString = "thisshouldberandom"
)

const htmlIndex = `<html><body>
Logged in with <a href="/login">facebook</a>
</body></html>
`

func handleMain(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "text/html; charset=utf-8")
  w.WriteHeader(http.StatusOK)
  w.Write([]byte(htmlIndex))
}

func handleFacebookLogin(w http.ResponseWriter, r *http.Request) {
  Url, err := url.Parse(oauthConf.Endpoint.AuthURL)
  if err != nil {
    log.Fatal("Parse: ", err)
  }
  parameters := url.Values{}
  parameters.Add("client_id", oauthConf.ClientID)
  parameters.Add("scope", strings.Join(oauthConf.Scopes, " "))
  parameters.Add("redirect_uri", oauthConf.RedirectURL)
  parameters.Add("response_type", "code")
  parameters.Add("state", oauthStateString)
  Url.RawQuery = parameters.Encode()
  url := Url.String()
  http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleFacebookCallback(w http.ResponseWriter, r *http.Request) {
  token := "EAAgpLQ9L0DQBADxiSRPSKpj8fJycgmm3xYhtldHPs7kL3LcVqjXdnswNQWKZAUl6osufzAejjALNlKYlq0Xb6WPeDZAbN9jF5uwIAu6TIUA2ya4y1ZBjm3yDJO52LySl0V903h4voZB4sGiPrBDDGoreI4arx8h0GSiShyJ4ENi2LSzHZBRMRNSr628VeimMMKOdwcK8yThc0KnyjcBI9tbsatbmxTmgZD"
  resp, err := http.Get("https://graph.facebook.com/me?fields=id,name,email&access_token="+token)
  if err != nil {
    fmt.Printf("Get: %s\n", err)
    http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
    return
  }
  defer resp.Body.Close()

  response, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    fmt.Printf("ReadAll: %s\n", err)
    http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
    return
  }

  log.Printf("parseResponseBody: %s\n", string(response))

  http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func main() {
  http.HandleFunc("/", handleMain)
  http.HandleFunc("/login", handleFacebookLogin)
  http.HandleFunc("/oauth2callback", handleFacebookCallback)
  fmt.Print("Started running on http://localhost:9090\n")
  log.Fatal(http.ListenAndServe(":9090", nil))
}