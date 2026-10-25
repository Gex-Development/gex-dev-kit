---
name: gex-local-spark-tests
description: Prepare runtime local PySpark 3.3 e Iceberg 1.0 para regressões comportamentais dos jobs canônicos do gex-infra-aws.
---

# Testes Spark locais do canônico

Use em checkout autorizado de `gex-infra-aws` para provar comportamento dos blocos reais de transformação e escrita Spark, não como substituto do aceite de produção. Skill não contém jobs, dados nem fixture: usa testes **do checkout atual**. Instalação da skill não instala Python, Java, JAR ou acessa AWS.

## Pré-requisitos do projeto

- Leia `AGENTS.md` e `.github/workflows/canonical-ci.yml` **do checkout** antes de instalar/executar. Se versões ou SHA mudaram, siga CI atual e não estas versões de referência.
- Python 3.10, `uv`, JDK 17, acesso temporário a Maven Central para JAR e espaço local para Spark. Runtime do CI consultado: `pyspark==3.3.0`, `pyarrow==21.*`, `iceberg-spark-runtime-3.3_2.12-1.0.0.jar` SHA-256 `ec879ae8f25f0996f4e212f844d6aea289bf0a7b29476b889d7042b12e0b06b4`.
- Confira conteúdo e efeitos do teste alvo antes de rodar: não execute teste que toque Glue, S3, Athena, recursos remotos ou clientes reais. Bloqueie credenciais AWS e serviço de metadata de instância; fixture Spark/Iceberg deve usar warehouse temporário e catálogo Hadoop local.

## Ambiente local descartável

A partir da raiz do `gex-infra-aws` e **somente depois de inspecionar o teste**:

```bash
WORK=$(mktemp -d)
uv venv -p 3.10 "$WORK/venv"
uv pip install -p "$WORK/venv" 'pyspark==3.3.0' 'pyarrow==21.*'
curl --fail --location --output "$WORK/iceberg.jar" 'https://repo.maven.apache.org/maven2/org/apache/iceberg/iceberg-spark-runtime-3.3_2.12/1.0.0/iceberg-spark-runtime-3.3_2.12-1.0.0.jar'
printf '%s  %s\n' ec879ae8f25f0996f4e212f844d6aea289bf0a7b29476b889d7042b12e0b06b4 "$WORK/iceberg.jar" | shasum -a 256 --check
```

Se `shasum` indisponível, confira mesmo SHA-256 com ferramenta confiável local; **não rode com JAR sem conferir**. Configure `JAVA_HOME` para JDK 17 no próprio terminal. Antes do primeiro SparkContext:

```bash
export SPARK_LOCAL_IP=127.0.0.1
export AWS_EC2_METADATA_DISABLED=true AWS_ACCESS_KEY_ID=local-test AWS_SECRET_ACCESS_KEY=local-test AWS_DEFAULT_REGION=us-east-1
export PYSPARK_PYTHON="$WORK/venv/bin/python" PYSPARK_DRIVER_PYTHON="$WORK/venv/bin/python"
export ICEBERG_JAR="$WORK/iceberg.jar"
export PYSPARK_SUBMIT_ARGS="--jars $WORK/iceberg.jar pyspark-shell"
"$WORK/venv/bin/python" -m unittest discover -s scripts/canonical/tests/spark -p 'test_*.py' -v
```

Não substitua o `-p` por nome fictício: use `test_*.py` somente após conferir todos os arquivos coletados, ou escolha `test_<arquivo_real>.py` inspecionado. Se o teste falhar por falta do JAR/runtime, reporte **BLOQUEADO**, não pule. Contraprova de mudança que escreve Iceberg inclui SQL real do job (MERGE/DELETE), replay e isolamento de gateway vizinho; AST isolada, mocks ou compilação não provam comportamento. Teste sintético local não é CHK histórico pós-reprocessamento nem validação em produção. Remova apenas `$WORK` criado nesta sessão após terminar; não limpe diretórios alheios.
