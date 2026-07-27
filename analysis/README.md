# Manual das funções da análise

Guia de estudo das funções que sustentam os notebooks 00–11. A ordem segue os blocos do
`00_setup.ipynb`; no fim entram as funções locais dos notebooks 06, 09 e 10. Cada entrada diz o
que a função faz, porque está feita assim e onde é que costuma enganar quem lê.

A unidade de análise é a **célula**, o par (avaliação, padrão);
cada avaliação tem 19 células, uma por padrão da taxonomia. `ann` é o dataframe das anotações
(`annotations.csv`), `gold` o das adjudicações (`gold.csv`), `status` o do estado de cada par
avaliação-anotador. Os booleanos vêm do exportador como as strings `"true"`/`"false"` e são
convertidos no carregamento.

---

## 1. Configuração e carregamento

### `ZIP` e `RUN` (constantes)
O `ZIP` aponta para a exportação analítica; o `RUN` escolhe a execução em análise (a 2 é a de
referência, a adjudicada). Tudo o resto lê estas duas constantes, pelo que mudar de execução é
mudar uma linha e correr de novo.

### `open_zip(filename=None)`
Abre o zip da exportação. Sem argumento, usa o `ZIP` configurado. Existe como função separada
para o carregamento nunca acontecer no import; só quando alguém pede dados.

### `_zip()`
Guarda o zip aberto num singleton. Ou seja, o ficheiro abre-se uma única vez por sessão,
independentemente de quantos `load` se fizerem. O prefixo `_` marca-a como interna.

### `load(name, run=REFERENCE_RUN)`
Lê um CSV do zip e devolve um dataframe. Três comportamentos a conhecer. Primeiro, tem cache
por par (ficheiro, run); pedir duas vezes o mesmo devolve o mesmo objecto sem tocar no disco.
Segundo, se o CSV tiver a coluna `runId` e o argumento `run` não for `None`, filtra logo para
essa execução; com `run=None` vêm todas as execuções, que é o que os notebooks 09 e 10 usam.
Terceiro, converte as colunas `present` e `finalLabel` de string para booleano. Armadilha
habitual, pedir `load("gold.csv")` e esquecer que o gold traz as duas passagens; quem quer só
uma tem de filtrar por `pass`.

---

## 2. Painel e regras de agregação

### `panel_majority(ann)`
Reduz os votos do painel a uma decisão por célula segundo a **maioria estrita relativa aos
membros que votaram**. Ou seja, presente quando mais de metade dos votos existentes são
positivos; num painel completo é 3 em 4, num painel parcial de três votos é 2 em 3. Um empate
de 2 contra 2 conta como ausente. É a regra que o ecrã de adjudicação mostrou, e por isso é a
regra de referência de toda a análise.

### `panel_vote(ann, k, label)`
A alternativa **absoluta**: presente onde pelo menos `k` modelos sinalizaram, seja qual for o
número de votantes. `k=1` é a união (qualquer modelo chega), `k=4` a unanimidade. A diferença
para a `panel_majority` só aparece nos painéis parciais; com 3 votantes, a maioria relativa
exige 2, mas a regra absoluta `k=3` exige 3, pelo que as duas divergem exactamente nessas
avaliações. É por isso que a maioria da tabela de calibração e a regra `>= 3` da análise de
sensibilidade não coincidem ao cêntimo.

---

## 3. Fiabilidade (concordância entre os modelos)

### `_matrix(ann, code)`
Prepara a matriz anotador × avaliação de um padrão, com 1/0/NaN (NaN onde o modelo não
respondeu). É o formato que os coeficientes de concordância esperam. Tudo o resto neste bloco
começa por aqui.

### `alpha(ann, code)`
O α de Krippendorff do padrão, via pacote `krippendorff`. Mede a concordância corrigida do
acaso, com o modelo de acaso estimado a partir da distribuição global dos valores. Devolve NaN
quando não há variância, ou seja, quando todos os modelos disseram sempre ausente; sem
desacordo possível, o coeficiente não está definido. Nos padrões raros o α cai mesmo com
concordância bruta altíssima; é o paradoxo da prevalência, e é a razão de o AC1 andar ao lado.

### `pairwise_kappa(ann, code)`
O κ de Cohen calculado para cada par de modelos e devolvido em média. Existe um caso degenerado
tratado à mão, quando os dois modelos do par deram sempre a mesma resposta única; o
`cohen_kappa_score` devolveria NaN e a função devolve 1, porque concordância total é
concordância total.

### `ac1(ann, code)`
O AC1 de Gwet, via pacote `irrCAC`. A diferença para o α está no modelo de acaso; o AC1 estima
uma prevalência única a partir da média das marginais de todos os anotadores, o que o mantém
estável sob desequilíbrio extremo de classes e robusto a enviesamentos individuais. Na prática,
nos padrões raros o α desaba e o AC1 fica perto de 1; a divergência entre os dois é informação,
não defeito.

### `specific_agreement(ann, code)`
A concordância específica positiva e negativa (P_pos, P_neg). A P_pos responde a uma pergunta
concreta: quando um modelo diz presente, com que frequência outro modelo também diz? É a medida
que melhor resiste nos padrões raros, porque olha só para os positivos. Uma P_pos baixa com
suporte razoável significa que os modelos raramente convergem nos mesmos positivos, ou seja,
desacordo genuíno na delimitação do padrão.

### `agreement_table(ann)`
Junta tudo numa tabela por padrão, com α, κ, AC1, P_pos, P_neg e o número de votos positivos,
ordenável pelo suporte. É a tabela de fiabilidade; a regra de leitura é nunca interpretar um
coeficiente sem olhar para o suporte ao lado.

---

## 4. Validade (calibração contra a referência) e direcção

### `cells(ann, gold, pass_="open")`
Constrói o dataframe de células que cruza a maioria do painel e cada modelo (`y_pred`) com o
rótulo adjudicado (`y_true`), restrito às avaliações adjudicadas na passagem escolhida. É a
ponte entre as anotações e a referência; toda a validade parte daqui.

### `calibration(cell_df)`
Precisão, revocação e F1 por par (modelo, padrão), com o suporte respectivo. A maioria do
painel entra como se fosse mais um modelo, o que permite compará-la com os membros na mesma
tabela.

### `macro(cal)`
A média macro por modelo, ou seja, a média dos F1 por padrão com peso igual para cada padrão.
É a convenção mais exigente, porque os padrões raros pesam tanto como os frequentes; um modelo
não sobe na macro por acertar muito no PM-1.

### `micro(cell_df)` (local do 06)
O F1 com todas as células agregadas num único cálculo, sem separar por padrão. Sobe muito face
à macro porque os padrões frequentes dominam. Reportar as duas granularidades é o que evita
leituras convenientes.

### `bootstrap_macro_f1(cell_df, n=2000, seed=0)`
O intervalo de confiança a 95% do F1 macro, por reamostragem. O desenho importa e convém
sabê-lo explicar. Reamostram-se **avaliações completas** com reposição, e não células, porque
as 19 células de uma avaliação partilham o texto e o adjudicador; reamostrar células fingiria
independência e daria intervalos estreitos de mais. São 2 000 réplicas com a semente do gerador
fixa em 0, para o intervalo ser reproduzível. O `merge` com `how="left"` é o truque que faz a
reposição funcionar, trazendo as células de cada avaliação escolhida tantas vezes quantas ela
saiu na reamostra. No fim tomam-se os percentis 2,5 e 97,5 da distribuição simulada.

### `adjudication_direction(ann, gold, pass_="open")`
Compara o rótulo final com a maioria do painel da execução em análise e classifica cada célula
como confirmação ou substituição, decompondo as substituições em adições (painel ausente, autor
presente) e remoções (o inverso). A direcção é **derivada** das anotações vivas, e não lida da
coluna armazenada, porque a coluna armazenada pertence à execução que estava no ecrã; derivar
permite o exercício hipotético sobre outras execuções, desde que assinalado como tal.

---

## 5. Saúde dos dados e modos de falha

### `classify(text)` e `FENCE`
Classifica cada resposta falhada num modo de falha (JSON inválido, truncamento, recusa, código
fora da taxonomia…), com uma expressão regular auxiliar para detectar blocos de código. Serve
para mostrar que as falhas são de formatação e não de conteúdo.

### `status_by_model(status)`
Conta, por modelo, os pares concluídos e falhados. É a primeira tabela de saúde; concentrações
num só modelo aparecem logo aqui.

### `partial_panel(status)`
Identifica as avaliações em que pelo menos um modelo não concluiu, os chamados painéis
parciais, e devolve a contagem e a percentagem. A maioria relativa lida com eles sem os
excluir.

### `adjudicated_on_partial(status, gold, pass_="open")`
Cruza os painéis parciais com as avaliações adjudicadas, para responder à pergunta que
interessa à validade: quantas células da referência assentam num painel incompleto? A análise
de sensibilidade que as exclui e recalcula mostra se são inócuas.

---

## 6. Figuras

### `styled_plt()`
Devolve o pyplot já configurado com o estilo sóbrio das figuras, tons de cinzento, sem grelha
pesada, sem molduras supérfluas. O import do matplotlib é tardio, para que o setup nunca falhe
num ambiente sem ele.

### `virgula(x)`
Formata números com vírgula decimal (0.45 passa a 0,45) nos eixos e rótulos.

### `save_fig(fig, name)`
Grava a figura em `figures/<name>.png` a 200 dpi com corte justo. As figuras são artefactos
derivados; regeneram-se sempre que o notebook corre.

---

## 7. Funções locais dos notebooks

### 09_factorial-2x2

- `macro_micro(pred, run, reviews)` e `run_scores(run, reviews)` — pontuam a maioria de uma
  execução (macro e micro) contra a referência, restringindo às avaliações adjudicadas; o
  `run_scores` acrescenta o melhor modelo isolado. São a base do quadro 2×2.
- `run_scores_models(run, models)` e `panel_shared(run)` — o mesmo, mas com o painel restrito a
  uma lista de modelos. Servem a comparação justa do painel comercial a 3 modelos, com o claude
  excluído dos dois lados.
- `model_delta(rA, rB)` — a diferença de F1 macro por modelo entre duas execuções com o mesmo
  painel. Isola o efeito do contexto dentro de cada painel; um modelo descontinuado aparece com
  delta NaN em vez de desaparecer em silêncio.
- `PR(run, mdl)` — precisão, revocação e contagem de sinalizações de um modelo numa execução.
  É a decomposição que mostra o mecanismo de supressão do contexto.
- `micro_blind(run, mdl)` — o F1 micro contra a referência cega, para o teste de circularidade;
  se o contexto degrada também contra uma referência que nunca viu painel nenhum, o efeito não
  é artefacto da semeadura.
- `version_delta(r_old, r_new)` — o salto v1→v2 por modelo, com painel e população fixos.
- `maj_series(run, reviews)` e `linhas_de(lab, vd)` — auxiliares da comparação de maiorias
  entre painéis e da montagem das tabelas.

### 10_comparison

- `f1_by_pattern(ann, gold, pass_)` — F1 por padrão de uma execução contra a referência; a
  comparação entre duas execuções é a diferença destas tabelas.
- `changed_knobs(a, b)` — lista o que difere entre as configurações de duas execuções (prompt,
  taxonomia, painel, contexto), ou seja, aquilo a que a diferença de F1 pode ser atribuída. O
  código recusa comparar execuções de populações diferentes.
- `majorities(ann)` — a maioria de cada painel por célula, para a divergência sem referência.
- `panel_verdict(ann, rule)` — a agregação com regra configurável, `strict` para a maioria
  relativa ou um inteiro para a regra absoluta. Existe porque comparar prevalências entre
  painéis de tamanhos diferentes só é justo com a mesma regra absoluta; a maioria estrita é um
  patamar mais baixo num painel de 3 do que num de 4.

### 08_anchoring

O intervalo de Wilson da proporção de divergência é calculado directamente na célula, sem
função própria; a fórmula fechada serve porque se trata de uma proporção simples (12 em 629),
e o z de 1,96 corresponde aos 95%.
