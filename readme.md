raspberry pi錠で、metabaseで描画するためのデータを収集する。
本コードは、センサからデータを取得し、SQLサーバに保存する。
別途起動したmetabaseは、このDBを参照して表示する。

開発メモはOneNoteに。

ビルド概要は下記
- ビルドおよび、raspberry piへの物件のコピーはtasks.jsonを使う
- 事前にsettings.jsonに、対象のパスを入力しておく