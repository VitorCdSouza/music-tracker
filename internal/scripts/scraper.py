import sys
import re
import json
from librespot.core import Session
from librespot import metadata
from google.protobuf.json_format import MessageToDict

def extract_playlist_id(url):
    match = re.search(r'playlist/([a-zA-Z0-9]+)', url)
    if match:
        return match.group(1)
    return None

def start_saved_session(creds_path):
    try:
        session = Session.Builder().stored_file(creds_path).create()
        return session
    except Exception as e:
        print(f"erro ao carregar sessao: {e}", file=sys.stderr)
        sys.exit(1)

def main(creds_path, playlist_url):
    playlist_id = extract_playlist_id(playlist_url)
    if not playlist_id:
        print("url invalida", file=sys.stderr)
        sys.exit(1)

    print("carregando credenciais e conectando ao spotify...")
    session = start_saved_session(creds_path)

    print("conectado, buscando playlist")

    try:
        playlist_id_obj = metadata.PlaylistId.from_base62(playlist_id)
        playlist_proto = session.api().get_playlist(playlist_id_obj)
        playlist_data = MessageToDict(playlist_proto, preserving_proto_field_name=True)

        playlist_name = playlist_data.get('attributes', {}).get('name', 'Unknown')
        items = playlist_data.get('contents', {}).get('items', [])

        print(f"playlist: {playlist_name} - n de musicas: {len(items)}\n")

        music_list = []
        count = 0
        
        for item in items:
            uri = item.get('uri', '')
            if not uri.startswith('spotify:track:'):
                continue

            track_id_base62 = uri.split(':')[-1]
            track_id_obj = metadata.TrackId.from_base62(track_id_base62)

            #track_proto = session.api().get_metadata_4_track(track_id_obj)
            #track_data = MessageToDict(track_proto, preserving_proto_field_name=True)

            #name = track_data.get('name', 'Unknown')
            #album = track_data.get('album', {}).get('name', 'Unknown')

            #artists_list = track_data.get('artist', [])
            #artists_names = ", ".join([a.get('name', '') for a in artists_list])

            music_list.append({
                #"name": name,
                #"artist": artists_names,
                #"album": album,
                "spotify_id": track_id_base62
            })

            count += 1

        print("\njson")
        json_result = json.dumps(music_list, indent=2, ensure_ascii=False)
        print(json_result)

    except Exception as e:
        print(f"erro consultando metadados: {e}")
        sys.exit(1)

if __name__ == '__main__':
    if len(sys.argv) < 3:
        print("chamada incorreta. ex: python3 scraper.py credentials.jso 'https://open.spotify.com/playlist/...'", file=sys.stderr)
        sys.exit(1)

    main(sys.argv[1], sys.argv[2])
