# AGENTS.md

## Objetivo

Este arquivo e o ponte de entrada para qualquer IA consiga atuar no projeto.

## Visão Geral do Sistema

O Sitema é um sistema de gerenciamento para Editoras, o objetivo é que ela possa fazer o gerenciamento de seus editais, manuscrito, produtos e autores.

O backend (este projeto) é feito em GO.lang e temos um projeto de migrations para criação das evoluções de banco que se encontra em `../database/migrations`.

Temos dois frontends que consomem a aplicação, sendo uma o front da editora  `../editora-frontend` (roda por padrão em localhost na porta 4201 ) e outro a da loja `..\loja-frontend` (roda por padrão na porta 4202).

Possuimos também um banco em PostgreSQL

## Visão desse Seviço

Autalemnte ele concentra o que sera dividido em dois serviços, que são o backend da Editora e o da Loja. (roda por padrão na porta 8081)

### Backend Go

Padrao dominante:

- `entity`
- `repository`
- `service`
- `routes`

Fluxo padrao:

```text
Route -> Service -> Repository -> PostgreSQL
```

Cuidados:

- manter contratos administrativos e publicos coerentes
- validar impacto em `entity`, `service`, `repository` e rotas quando adicionar campos
- revisar `Scan`, placeholders SQL e JSON persistido ao alterar queries grandes
- quando mexer em configuracao da loja, considerar os fluxos de rascunho e publicacao


Backend:

```bash
cd editora/backend
go run .
```

Infra local:

```bash
docker compose up -d
```


### Portas padrao

- Admin: `http://localhost:4201`
- Storefront: `http://localhost:4202`
- Backend: `http://localhost:8081`
- PostgreSQL: `5432`


## Como pensar mudancas neste projeto

### Se a tarefa envolver configuracao da loja

Revise sempre:

- admin de layout em `editora-frontend/src/app/features/loja`
- backend da loja em `editora- backend/src/loja`
- storefront em `loja-frontend/src/app/features/store`

Mudancas de configuracao normalmente exigem alinhamento entre:

- payload TypeScript
- entidade Go
- service Go
- repository Go
- resposta publica da loja
- aplicacao visual no storefront


## Checklist para qualquer agente antes de finalizar

- entendeu se a mudanca impacta admin, backend, storefront ou mais de uma camada
- revisou contratos e payloads dos dois lados quando adicionou campo novo
- criou migration quando houve alteracao estrutural de banco
- preservou as regras de UI do admin e Loja
- manteve texto da interface em portugues humano
- nao quebrou o comportamento de rascunho/publicacao da loja
- validou arquivos alterados com diagnostico
- rodou build/teste relevante quando a mudanca foi substantiva

## Resumo rapido para agentes

Se voce acabou de chegar neste projeto, assuma o seguinte:

- e um monorepo com backend Go, admin Angular e storefront Angular
- a loja e altamente configuravel e depende de contratos entre admin, backend e frontend publico
- livro e produto
- bootstrap, dark theme e mobile first sao obrigatorios
- no admin nao pode haver inline template/style
- componentes devem ser reaproveitados antes de duplicar
- tipografia, cores, banners, cards e integracoes da loja sao areas sensiveis
- salvar integracoes nao deve publicar o resto do rascunho da loja