package hello

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sync"

	"github.com/ChimeraCoder/anaconda"
	"github.com/mrjones/oauth"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

var (
	TwitterConsumerKey    string
	TwitterConsumerSecret string
	AuthCallbackUrl       string
)

func init() {
	TwitterConsumerKey = os.Getenv("TWITTER_CONSUMER_KEY")
	TwitterConsumerSecret = os.Getenv("TWITTER_CONSUMER_SECRET")
	AuthCallbackUrlBase := os.Getenv("TWITTER_CALLBACK_URL_BASE")
	AuthCallbackUrlRel := "/auth/callback"
	AuthCallbackUrl = fmt.Sprint(AuthCallbackUrlBase, AuthCallbackUrlRel)

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/auth", authHandler)
	http.HandleFunc(AuthCallbackUrlRel, authCallbackHandler)
	http.HandleFunc("/omikuji", omikujiHandler)
}

var rootTemplate = template.Must(template.New("root").Parse(`
<!Doctype HTML>
<html>
	<head>
		<meta charset=utf-8>
		<title></title>
	</head>
	<body>
		<a href="/auth">ツイッターログイン</a>
		<a href="/omikuji">おみくじ</a>
	</body>
</html>
`))

func rootHandler(w http.ResponseWriter, r *http.Request) {
	rootTemplate.Execute(w, nil)
}

func init() {
	anaconda.SetConsumerKey(TwitterConsumerKey)
	anaconda.SetConsumerSecret(TwitterConsumerSecret)
}

var omikujiTemplate = template.Must(template.New("omikuji").Parse(`
<!Doctype HTML>
<html>
	<head>
		<meta charset=utf-8>
		<title></title>
	</head>
	<body>
		 大吉です!
	</body>
</html>
`))

func omikujiHandler(w http.ResponseWriter, r *http.Request) {
	omikujiTemplate.Execute(w, nil)
}

var (
	RequestTokenMu sync.Mutex
	RequestTokens  []oauth.RequestToken
)

func AuthRequest(rtoken oauth.RequestToken) {
	RequestTokenMu.Lock()
	defer RequestTokenMu.Unlock()
	RequestTokens = append(RequestTokens, rtoken)
}

func FindRequest(ctx context.Context, token string) *oauth.RequestToken {
	RequestTokenMu.Lock()
	defer RequestTokenMu.Unlock()
	for _, rtoken := range RequestTokens {
		log.Debugf(ctx, "token:%v", rtoken.Token)
		if rtoken.Token == token {
			r := rtoken
			return &r
		}
	}
	return nil
}

func NewTwitterClient(ctx context.Context) *oauth.Consumer {
	httpClient := urlfetch.Client(ctx)
	c := oauth.NewCustomHttpClientConsumer(
		TwitterConsumerKey, TwitterConsumerSecret,
		oauth.ServiceProvider{
			RequestTokenUrl:   "https://api.twitter.com/oauth/request_token",
			AuthorizeTokenUrl: "https://api.twitter.com/oauth/authorize",
			AccessTokenUrl:    "https://api.twitter.com/oauth/access_token",
		}, httpClient)
	return c
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	c := NewTwitterClient(ctx)
	requestToken, u, err := c.GetRequestTokenAndUrl(AuthCallbackUrl)
	log.Debugf(ctx, "requestToken:%v,%v callbackUrl:%v", requestToken.Token, requestToken.Secret, AuthCallbackUrl)
	if err != nil {
		fmt.Fprint(w, err)
		w.WriteHeader(500)
		return
	}
	AuthRequest(*requestToken)
	http.Redirect(w, r, u, http.StatusFound)
}

func authCallbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	c := NewTwitterClient(ctx)
	token := r.FormValue("oauth_token")
	verificationCode := r.FormValue("oauth_verifier")
	log.Debugf(ctx, "oauth_token:%v oauth_verifier:%v", token, verificationCode)
	requestToken := FindRequest(ctx, token)
	if requestToken == nil {
		fmt.Fprint(w, "token not found")
		w.WriteHeader(500)
		return
	}
	accessToken, err := c.AuthorizeToken(requestToken, verificationCode)
	if err != nil {
		fmt.Fprint(w, err)
		w.WriteHeader(500)
		return
	}
	fmt.Fprintln(w, accessToken.AdditionalData)
	api := anaconda.NewTwitterApi(accessToken.Token, accessToken.Secret)
	api.HttpClient = urlfetch.Client(ctx)
	ok, err := api.VerifyCredentials()
	if err != nil {
		fmt.Fprint(w, err)
		w.WriteHeader(500)
		return
	}
	fmt.Fprintln(w, ok)
}
