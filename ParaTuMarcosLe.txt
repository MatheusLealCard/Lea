1 - Baixa o docker desktop "https://docs.docker.com/desktop/setup/install/windows-install/"
2 - Baixa o go "https://go.dev/doc/install"
3 - No terminal usa o comando "go mod tidy"
4 - entra no terminal na pasta chamada "docker" e usa o comando "docker compose up -d"
5 - na pasta raiz abre o terminal e roda go run main.go

Pastas:
    database:
        onde fica as config do banco, se adicionar algo na entities 
        tem que adicionar aqui, lá no "DB.AutoMigrate"

    entities: 
        onde fica as tables do banco

    handlers:
        onde ficam as funcoes (get, delete, create etc)

Ta sem o .env vou te mandar pra tu criar na raiz da pasta, cria um arquivo chamado ".env".

Quando tu terminar de fazer o que tu tiver fazendo ou quiser mandar pro git tu faz: 
    1 - git checkout -b "nome que tu quiser(mas poe algo relacionado com o que tu ta fazendo ex: loginHandler)"
    2 - git add .
    3 - git commit -m "explciacao do que tu fez resumido, no max 10 palavras"
    4 - git push origin nome-que-tu-criou-no-checkout

    (Se ja tiver criado a sua brach comeca do passo 2)
