# Blockchain PoS em Go

Projeto de aprendizado para construir uma blockchain Proof of Stake em Go, do zero, com camadas separadas:

- `internal/chain`: regras de blockchain, consenso, staking, assinaturas e slashing.
- `internal/wallet`: carteiras ECDSA e verificação de assinaturas.
- `internal/p2p`: mensagens HTTP simples entre nós.
- `cmd/node`: composição das camadas, mempool, produção de blocos e tratamento de mensagens P2P.

## Roadmap

- [x] Estruturas base: blocos, transações e hash SHA-256.
- [x] Chain local single-node: `AddBlock` e `IsValid`.
- [x] Carteiras e assinaturas de transações com ECDSA.
- [x] Staking e seleção determinística ponderada por stake.
- [x] Rede P2P HTTP simples com mempool e broadcast.
- [x] Slashing simplificado para double-signing de blocos conflitantes.

## Slashing

Esta etapa é opcional e exploratória. O objetivo é entender o mecanismo, não produzir um protocolo pronto para produção.

O slashing implementado detecta equivocação/double-signing quando existem dois blocos:

- com o mesmo índice;
- apontando para o mesmo `PrevHash`;
- assinados pelo mesmo `Validator`;
- com hashes diferentes.

Quando a evidência é válida, o stake do validador culpado é zerado no `ValidatorSet`. Como `SelectValidator` usa o stake como peso, esse validador deixa de ser escolhido nas próximas rodadas enquanto estiver com stake zero.

### Evidência

A evidência é representada por `chain.SlashingEvidence`, contendo os dois blocos conflitantes. Cada bloco agora possui assinatura explícita (`Signature`), assinada pela carteira do validador sobre o hash do bloco.

A assinatura do bloco não entra no cálculo do hash. Isso evita dependência circular:

1. calcula-se o hash com os dados do bloco;
2. assina-se esse hash;
3. verifica-se primeiro o hash, depois a assinatura.

Sem assinatura explícita no bloco, a evidência seria fraca: um nó malicioso poderia fabricar um bloco com `Validator` apontando para outra carteira. Com assinatura, ele ainda pode encaminhar evidência, mas não consegue forjar a prova sem a chave privada do validador acusado.

## Limitações

Esta versão não implementa recursos de blockchains de produção, como:

- período de unbonding;
- disputa formal da evidência;
- snapshots históricos de stake;
- verificação de que o validador acusado era o selecionado naquele height;
- persistência de evidências em disco.

## Teste prático

Para testar manualmente a ideia, rode a rede com:

```bash
docker compose up --build
```

Depois simule um validador desonesto produzindo dois blocos diferentes no mesmo índice e com o mesmo `PrevHash`, assinados pela mesma carteira, e envie cada bloco para peers diferentes via `/block`. Quando um nó observar os dois blocos conflitantes, ele deve:

1. validar as assinaturas dos dois blocos;
2. criar uma `SlashingEvidence`;
3. aplicar a penalidade no `ValidatorSet`;
4. propagar a evidência via `/slashing`.

Nos logs, procure mensagens como:

```text
Slashing aplicado: stake do validador ... foi zerado
Slashing recebido via P2P e aplicado contra validador ...
```

## Validação

```bash
GOCACHE=/tmp/go-build-cache go test ./...
docker compose config
```
