# music-tracker

Um app de terminal (TUI) para buscar e baixar músicas ou playlists do Spotify direto para o seu PC. 

A interface é toda feita em **Go** (usando [Bubble Tea](https://github.com/charmbracelet/bubbletea)) e a engine de download roda por baixo dos panos em **Python** (usando a lib `librespot` para conectar direto nos servidores do Spotify).

## Funcionalidades

- **Baixa em alta qualidade**: Pega o stream direto em `.ogg` (Vorbis). Tenta baixar em `very_high` (se você tiver Premium) ou faz fallback automático pra `high`.
- **Metadados**: Já injeta as tags de artista, nome da música e ID do Spotify no arquivo de áudio.
- **Detecção de duplicatas**: Se a música já existir na pasta, ele pula pra não baixar de novo.
- **Gerenciador de arquivos embutido**: Dá pra navegar pelas pastas de download direto pelo terminal.
- **Login OAuth**: Autenticação fácil pelo navegador, salvando a sessão num `credentials.json` local.

## Pré-requisitos

Você vai precisar ter instalado na máquina:
- **Go** (1.20+)
- **Python** (3.8+)

Além disso, instale as dependências do Python necessárias para os workers funcionarem:

```bash
pip install librespot mutagen protobuf
```
*(Nota: dependendo do seu SO, talvez seja necessário usar `pip3`)*

## Como rodar

1. Clone o repositório:
```bash
git clone https://github.com/vitorcds/music-tracker.git
cd music-tracker
```

2. Baixe as dependências do Go:
```bash
go mod tidy
```

3. Rode o app:
```bash
cd cmd/tracker
go run main.go
```

## Como usar (Atalhos)

A interface tem dois modos (Normal e Input), parecido com o Vim.

- `i`: Entra no modo de digitação (Input mode).
- `Esc`: Sai do modo de digitação (Normal mode).
- `H` e `L`: Navega entre as abas (Busca, Arquivos, Configuração).
- `j` e `k`: Navega nas listas (arquivos, opções de menu).
- `Enter`: Confirma ações / Inicia o download.
- `Ctrl+S`: Salva as configurações.
- `Ctrl+C`: Sai do app.

Na **primeira execução**, o app vai pedir para você se autenticar. Ele vai tentar abrir o seu navegador para fazer login no Spotify. Depois do login, ele salva um arquivo `credentials.json` para os próximos acessos.

Para baixar, basta colar o link de uma música ou playlist na aba de busca (ex: `https://open.spotify.com/playlist/...`) e apertar Enter.

## Estrutura do projeto

- `cmd/tracker/`: Ponto de entrada do app em Go.
- `internal/ui/`: Telas e componentes do terminal (Bubble Tea).
- `internal/bridge/`: Comunicação (IPC) entre o frontend em Go e os scripts em Python.
- `internal/scripts/`: Scripts em Python (`scraper.py`, `downloader.py`, `worker.py`) que lidam com a API do Spotify via `librespot`.
- `internal/config/`: Gerenciamento do arquivo de configurações (`config.json`).

## TODO / Roadmap

- [ ] Suporte para downloads do YouTube (estrutura já iniciada no código).
- [ ] Melhorar o feedback de progresso dos downloads na interface.
- [ ] Refatorar a comunicação de eventos Python -> Go.

---
**Aviso:** Este projeto é apenas para fins educacionais. Respeite os termos de serviço das plataformas e os direitos autorais dos artistas.
