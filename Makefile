.PHONY: lab play

# Variables
CEP=73340608

up:
	docker compose up -d;

down:
	docker compose down;

restart:
	docker compose restart;


s-a:
	@sleep 10s ;
	curl -X POST -d '{"cep": "$(CEP)"}' http://localhost:8080
	@echo '\n' ;

s-b:	
	@sleep 10s ;
	curl http://localhost:8081/weather?cep=$(CEP)
	@sleep 10s ;
	@echo '\n' ;

services: up s-a s-b

request: s-a s-b