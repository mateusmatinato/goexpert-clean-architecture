### 
# Exercício 3 - Clean Architecture

### Introdução
Exercício da Pós Graduação em Desenvolvimento Avançado com Golang.

### Para executar
Clone o repositório e acesse a pasta pelo terminal.

Execute o seguinte comando:
```bash
go run cmd/ordersystem/main.go cmd/ordersystem/wire_gen.go
```

### Observações
A base de dados utilizada para armazenar as orders foi o sqlite3. 
Ao subir a aplicação, as migrations são executadas automaticamente.
Na pasta api, existem dois arquivos .http que podem ser utilizados para testar as chamadas REST para a aplicação.
No endpoint localhost:8080 é possível acessar o playground do GraphQL.
É possível testar a chamada gRPC utilizando o evans (ou alguma outra biblioteca gRPC de preferência).

### Instruções do exercício
Pra este desafio, você precisará criar o usecase de listagem das orders.
Esta listagem precisa ser feita com:
- Endpoint REST (GET /order)
- Service ListOrders com GRPC
- Query ListOrders GraphQL
Não esqueça de criar as migrações necessárias e o arquivo api.http com a request para criar e listar as orders.