# Agones GameServer Delete With Controller Down

このリポジトリは Agones の GameServer を削除するときの挙動について検証しまとめたものです。

Agones の GameServer は以下の2通りの方法で消すことができます。

1. `kubectl delete gs` などの delete API を呼ぶことで直接削除する
2. `kubectl edit gs` などで GameServer の STATE を Shutdown にすることで Agones Controller に削除してもらう

Agones Controller が正常に動作しているときやダウンしているときに、これらの方法で GameServer を削除したときの挙動について検証します。
また、削除の処理はシェル上で `kubectl` を用いたものと Go のプログラムによる削除の2通りを検証します。

## Prerequirements

* minikube
* virtualbox
  * minikube の driver として virtualbox を使います
* kubectl
* helm

## 環境構築

`make up` で minikube を用いた Agones の環境構築、 `make down` で minikube 環境の削除ができます。

## 検証に必要なリソース操作方法のまとめ

実際に検証に入る前に、検証に必要なリソースの操作方法をまとめます。

### Controller を上げ下げする方法

`kubectl scale` によってレプリカの数を変えます。
Controller を立ち上げたいときは1、 Controller を落としたいときは0にします。

```
$ kubectl scale -n agones-system deployment/agones-controller --replicas=0
```

### Fleet を立てる方法

今回の検証では簡単のため、 [公式のチュートリアル](https://agones.dev/site/docs/getting-started/create-gameserver/) に従って simple-game-server の Fleet を立てます。
また、 Fleet は `default` namespace に立てることとします。
以下のコマンドで Fleet を立てることができます。

```
$ kubectl apply -f https://raw.githubusercontent.com/googleforgames/agones/release-1.13.0/examples/simple-game-server/fleet.yaml
```

### GameServer や Pod の状態を目視で監視する

今回の検証では GameServer を削除しようとしたときの GameServer や Pod の挙動を確認する必要があります。
確認方法の1例として、 `watch` コマンドを用いた確認方法を紹介します。
`watch` は一定時間ごとにコマンドを実行し、出力をリフレッシュして出してくれるコマンドです。

以下のコマンドでは 1秒ごとに GameServer や Pod の状態を確認できます。

```
$ watch -n 1 kubectl get [pod / gs]
```

また、 `kubectl get gs -w` のように `-w` オプションをつけても似たようなことができますが、GameServer や Pod が消えたことを確認しづらいので watch コマンドを用いるほうがやりやすいと思います。

## 検証

準備が整ったので、実際に検証していきます。

### 直接消した場合

本節では冒頭で説明した2通りの消し方のうち、1つ目の delete API で直接消した場合について検証します。

#### シェル経由で消した場合

まず、 `kubectl delete gs [NAME]` で削除したときの挙動を説明します。

Controller が動いている場合は GameServer が削除されて新しい GameServer が作成されます。
このとき、削除される GameServer は対応する Pod の STATE が Terminating などになっている間や削除される直前まで STATE が Ready のままでした。

Controller がダウンしている場合は `kubectl` がブロックしたままになりプロンプトが返ってきません。
Controller を立ち上げ直すとそのうち削除が実行され、プロンプトが返ってきます。
立ち上げ直したときも削除される GameServer は直前まで Ready のままでした。

#### Go のプログラム経由で消した場合

次に、Go のプログラム経由で消した場合の挙動を説明します。
使用したプログラムは [down_by_delete/main.go](./down_by_delete/main.go) です。

Controller が動いている場合は GameServer が削除されて新しい GameServer が作成されます。
このとき、シェル経由で削除した場合と同様に、削除される GameServer は削除される直前まで STATE が Ready のままでした。

一方、Controller がダウンしている場合はシェル経由の場合と挙動が異なり、処理がブロックされることはありませんでした。
Controller がダウンしている間は GameServer や Pod の STATE が変化することはありませんでした。
Controller を立ち上げ直すとそのうち削除が実行されます。
立ち上げ直したときも削除される GameServer は直前まで Ready のままでした。

### STATE を Shutdown にした場合

本節では冒頭で説明した2通りの消し方のうち、2つ目の STATE を Shutdown にして Controller に削除してもらう場合について説明します。

#### シェル経由で消した場合

まず、 `kubectl edit gs [NAME]` で STATE を Shutdown にしたときの挙動を説明します。

Controller が動いている場合は GameServer の STATE が Shutdown になった後、 GameServer が削除され新しい GameServer が作成されます。
`kubectl delete gs` のときとの違いは、 STATE が Shutdown になった点です。

Controller がダウンしている場合は GameServer の STATE が Shutdown になった以降は状態変化が起きませんでした。
Controller を立ち上げ直すとそのうち削除が実行されます。

#### Go のプログラム経由で消した場合

次に、Go のプログラム経由で消した場合の挙動を説明します。
使用したプログラムは [down_by_update/main.go](./down_by_update/main.go) です。
直接削除した場合と異なり、シェル経由で消した場合と同じ挙動を示しました。

Controller が動いている場合は GameServer の STATE が Shutdown になった後、 GameServer が削除され新しい GameServer が作成されます。
直接削除したときとの違いは、 STATE が Shutdown になった点です。

Controller がダウンしている場合は GameServer の STATE が Shutdown になった以降は状態変化が起きませんでした。
Controller を立ち上げ直すとそのうち削除が実行されます。

## 考察

`kubectl delete gs` などで直接削除した場合と `kubectl edit gs` などで STATE を書き換えた場合の差は GameServer の STATE が更新されるかであるといえます。
直接削除した場合では STATE が Ready のままなので、本当は消えるはずの GameServer が直前まで使用可能な状態に見えてしまいます。
そのため、可能であれば STATE を書き換えたほうが終了状態を扱いやすくなると考えられます。

また、 Controller がダウンしていたときも、 STATE を書き換える手法では STATE が書き換わるため終了させるつもりの GameServer であることが外から判別できます。
このことからも、直接削除するよりも STATE を書き換えたほうが扱いやすいといえます。