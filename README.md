
# Pos Go Weather Tracing

Projeto desenvolvido em Go que receba um CEP, identifica a cidade e retorna o clima atual (temperatura em graus celsius, fahrenheit e kelvin) juntamente com a cidade. Esse sistema deverá implementar OTEL(Open Telemetry) e Zipkin.

## Processos

- Receber um CEP válido de 8 dígitos.
- Busca a localização de acordo com o CEP na API ViaCEP.
- Utilizar a API WeatherAPI para consultar a temperatura na localização encontrada.
- Retornar a temperatura nos formatos Celsius, Fahrenheit e Kelvin.
- Tracing distribuído com OpenTelemetry para facilitar a análise de desempenho.

## Como executar o projeto 🚀

### Subindo os serviços

1. Utilize o comando a seguir para subir os serviços e executar as atividades:

```bash
make services
```
2. Utilize o comando a seguir para mais requisiçoes:

```bash
make request
```

3. No seu navegador local, abra a URL e valide as evidências. Abaixo algumas imagens de referência:

http://localhost:9411/


### Destruindo os serviços
Para parar e remover os containers criados, use:
```bash
make down
```

### Limpando recursos Docker/Podman
Para remover containers, imagens e volumes não utilizados, execute:
```bash
make clean
```