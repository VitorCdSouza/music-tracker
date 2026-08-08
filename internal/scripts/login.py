import textwrap
import threading
import webbrowser

from librespot.core import MercuryRequests, OAuth, Session

SCOPES = [
    "streaming",
    "playlist-read-private",
    "playlist-read-collaborative",
    "user-follow-read",
    "user-read-playback-position",
    "user-top-read",
    "user-read-recently-played",
    "user-library-read",
    "user-read-email",
    "user-read-private",
]


def login_oauth(send_event, caminho_creds="credentials.json"):
    send_event("log", "autenticacao spotify")

    port = 4381
    redirect_url = f"http://127.0.0.1:{port}/login"

    def oauth_print(url):
        # magic_link = f"\033]8;;{url}\033\\ [ login no navegador ]\033]8;;\033\\"
        # print(f"\n{magic_link}\n")
        send_event("log", "abrindo navegador...")
        try:
            webbrowser.open(url)
        except Exception:  # noqa: BLE001, S110
            pass

        def print_fallback():
            send_event("log", "se navegador nao abriu, copie link abaixo:")
            wrapped_url = textwrap.fill(url, width=70, break_long_words=True)
            send_event("log", f"{wrapped_url}\n")

        t = threading.Timer(5.0, print_fallback)
        t.daemon = True
        t.start()

    try:
        client_id = MercuryRequests.keymaster_client_id
        oauth = (
            OAuth(client_id, redirect_url, oauth_print)
            .set_scopes(SCOPES)
            .set_listen_all(True)
        )
        login_credentials = oauth.flow()

        send_event("log", "processando token de login")

        builder = Session.Builder()
        builder.conf.store_credentials = True
        builder.conf.stored_credentials_file = caminho_creds
        builder.login_credentials = login_credentials

        _ = builder.create()

        send_event("log", "login realizado")
        send_event("auth_done", f"credentiais gravadar em: {caminho_creds}")

    except Exception as e:  # noqa: BLE001
        send_event("error", f"falha ao criar credentials.json: {e}")
