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

Migração para arquitetura de worker: um único `worker.py` vivo lendo comandos JSON pelo stdin,
no lugar de um processo Python novo por ação. Auth é a primeira ação migrada.

### Autenticação (Go ↔ worker)

- [x] Ajustar chamada do `worker.py` para o login (`login.py`)
- [x] Trocar os `print()` do `login.py` por eventos JSON — hoje o Go cai no fallback `Event: "log"` (`spotify.go:73`)
- [x] Tirar o `sys.exit(1)` do `login.py:64` — mata o worker inteiro depois do import
- [x] Rodar `oauth.flow()` fora da thread principal — bloqueia o loop do stdin (`login.py:52`)
- [ ] Mover as mensagens tipadas (`LineMsg`, `AuthDoneMsg`, `ScrapDoneMsg`, `DownloadDoneMsg`) para `interface.go` — são contrato do pacote `bridge`, não implementação do Spotify (`spotify.go:20`)
- [ ] Marcar as mensagens do worker com uma interface `WorkerMsg` (método vazio, minúsculo) — deixa o `Update` rearmar o `ListenForEvents` num ponto só, em vez de um `append` por case (`app.go:174`)
- [ ] Traduzir `Event: "auth_done"` em `AuthDoneMsg` dentro de `ListenForEvents` — hoje o `Update` compara string crua (`interface.go:36`, `app.go:167`)
- [ ] Remover o `if msg.Event == "auth_done"` vazio do ramo de erro do `json.Unmarshal` — o `msg` ali é zero-value, nunca dispara (`spotify.go:73`)
- [ ] Sair da tela de auth ao concluir — `AuthModel.state` trava em `1` (`auth.go:70`)
- [ ] Tratar retorno de `SpotifyProvider.Auth()` no `Update` do `AppModel` (`spotify.go:107`)
- [ ] Proteger provider `nil` — escolher YouTube chama `Auth()` sobre `nil` (`app.go:190`)
- [ ] Usar o ponteiro `provider` no `AuthModel` ou removê-lo (`auth.go:22`)
- [ ] Corrigir `login.py:68` — chama `login_oauth()` sem passar o `creds_path` do `argv`
- [x] Conferir `credentials.json` no `.gitignore` — contém token de sessão
- [ ] Destrackear os `.pyc` de `internal/scripts/__pycache__/` — já estão no índice do git

### Depois da autenticação

- [ ] Migrar `Scrap` e `Download` para o worker — ainda abrem processo separado (`spotify.go:123` e `:265`)
- [ ] Completar o diff online vs local no `Scrap()` — retorna `nil` quando há IDs locais (`spotify.go:254`)
- [ ] Disparar `Scrap` no Enter da busca — hoje só marca `downloading = true` (`app.go:210`)

### Revisar no final (caminhos)

Adiado de propósito: mexer aqui muda cwd e quebra tudo que depende de caminho relativo.
Fazer de uma vez só, depois que auth, scrap e download estiverem funcionando.

- [ ] Worker sobe com caminho relativo `../../internal/scripts/worker.py`, assumindo cwd em `cmd/tracker` (`spotify.go:36`)
- [ ] `HasCredentials()` procura `credentials.json` no cwd (`interface.go:56`)
- [ ] `scraper.py` e `downloader.py` recebem `"credentials.json"` hardcoded (`spotify.go:125` e `:267`)
- [ ] Import do `worker.py` é implicitamente relativo (`from login import ...`) — só resolve porque o script roda de dentro de `internal/scripts/`. Decidir entre `PYTHONPATH`/`-m` ou manter
- [ ] `login.py:20` tem default `credentials.json` relativo ao cwd do worker

### Robustez

Achados ao construir o music-player, que esbarra nos mesmos problemas.

- [ ] Descartar a sessão do librespot quando ela morre — `start_saved_session()` faz `sys.exit(1)` em qualquer falha e não tenta reconectar (`scraper.py:14`, `downloader.py:14`). O access point do Spotify recusa conexão de vez em quando (`Errno 111`), e reconectar resolve. Piora depois da migração: com o worker vivo, uma sessão morta fica em cache até o app reiniciar
- [ ] Parar o worker antigo antes de subir outro — `ConfigSavedMsg` chama `StartWorker` sem encerrar o anterior, deixando dois processos Python com duas sessões nas mesmas credenciais (`app.go:189`)
- [ ] Guardar o `*exec.Cmd` no `SpotifyProvider` — hoje só o stdin é guardado, então não há como matar o worker (`spotify.go:18`)
- [ ] Checar lista vazia antes do `enter` — `model.files[model.cursor]` estoura numa pasta sem arquivos (`files.go:45`)
- [ ] Baixar para `.part` e renomear no fim — download interrompido deixa `.ogg` truncado e sem tag, que a detecção de duplicata não reconhece e vira cópia nova na próxima tentativa (`downloader.py:114`)
- [ ] Tratar os erros engolidos — `scanner.Err()` e `cmd.Wait()` dentro de `if` vazio escondem falha do Python (`spotify.go:60`, `:80`, `:164`, `:191`, `:297`, `:299`)
- [ ] Resolver a duplicação entre `downloading` e `screenDownloading` — são dois campos para o mesmo estado, e a `View` chega a desenhar `download.View()` duas vezes (`app.go:33` e `:282`)

### Ideias posteriores

- [ ] Suporte para downloads do YouTube (estrutura já iniciada no código).
- [ ] Melhorar o feedback de progresso dos downloads na interface.
- [ ] Refatorar a comunicação de eventos Python -> Go.
- [ ] Busca por texto na aba de busca — hoje ela só aceita URL, apesar do README prometer busca. Exige a Web API (`/v1/search`), que precisa de um app próprio no dashboard: o token que o librespot carrega vem de um `client_id` compartilhado por todos os usuários da lib e devolve `429 Too Many Requests` já na primeira chamada, com `Retry-After` de dezenas de segundos. Os endpoints que o projeto usa hoje (`session.api().get_playlist()`, `get_metadata_4_track()`) não passam por aí e não têm esse problema
- [ ] Estilizar a TUI com lipgloss — última prioridade, só depois de auth, scrap e download funcionando

---
**Aviso:** Este projeto é apenas para fins educacionais. Respeite os termos de serviço das plataformas e os direitos autorais dos artistas.
