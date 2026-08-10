import json
import sys
import threading
import time

from login import login_oauth


def send_event(event, message):
    print(json.dumps({"event": event, "message": message}), flush=True)


def ping():
    while True:
        time.sleep(5)
        send_event("log", "worker ativo")


def main():
    send_event("log", "worker iniciado e aguardando comandos")

    t = threading.Thread(target=ping, daemon=True)
    t.start()

    for line in sys.stdin:
        line = line.strip()
        if not line:
            continue

        try:
            command = json.loads(line)
            action = command.get("action", "unknown")
            if action == "login":
                send_event("log", f"comando recebido do go: {action}")
                login_oauth(send_event)

        except Exception as e:  # noqa: BLE001
            send_event("error", f"falha no worker: {e}")


if __name__ == "__main__":
    main()
