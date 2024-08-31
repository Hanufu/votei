# Nome do binário
BINARY_NAME = votei

# Diretório onde o código-fonte principal está
SRC_DIR = ./cmd/votei

# Comando para compilar o binário
build:
	go build -o $(BINARY_NAME) $(SRC_DIR)

# Comando para rodar o binário
run: build
	./$(BINARY_NAME)

# Comando para rodar o código diretamente com go run
dev:
	go run $(SRC_DIR)

# Comando para limpar arquivos gerados
clean:
	rm -f $(BINARY_NAME)

# Comando para exibir a ajuda
help:
	@echo "Comandos disponíveis:"
	@echo "  make build   - Compila o binário"
	@echo "  make run     - Compila e executa o binário"
	@echo "  make dev     - Executa o código diretamente com go run"
	@echo "  make clean   - Remove o binário gerado"
	@echo "  make help    - Exibe esta mensagem de ajuda"
