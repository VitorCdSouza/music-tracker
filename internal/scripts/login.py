import sys
from librespot.core import Session, OAuth, MercuryRequests

SCOPES = [
    'streaming',
    'playlist-read-private',
    'playlist-read-collaborative',
    'user-follow-read',
    'user-read-playback-position',
    'user-top-read',
    'user-read-recently-played',
    'user-library-read',
    'user-read-email',
    'user-read-private'
]

def realizar_login_oauth(caminho_creds = "credentials.json"):
    print("script python - autenticacao librespot")

    port = 4381
    redirect_url = f"http://127.0.0.1:{port}/login"
    
    def oauth_print (url):
        print("link para login:")
        print(f"\n{url}\n")

    try:
        client_id = MercuryRequests.keymaster_client_id
        oauth = OAuth(client_id, redirect_url, oauth_print).set_scopes(SCOPES).set_listen_all(True)
        login_credentials = oauth.flow()

        print("processando token de login")

        builder = Session.Builder()
        builder.conf.store_credentials = True
        builder.conf.store_credentials_file = caminho_creds
        builder.login_credentials = login_credentials

        session = builder.create()

        print(f"login realizado")
        print(f"credentiais gravadar em: {caminho_creds}")
    
    except Exception as e:
        print(f"falha: {e}")
        sys.exit(1)

if __name__ == '__main__':
    creds_path = sys.argv[1] if len(sys.argv) > 1 else "credentials.json"
    realizar_login_oauth()

