# go install github.com/aggronmagi/gocc@latest
# mv ~/go/bin/gocc ~/go/bin/gocc-walle 
 
gocc-walle -o protobuf protobuf.bnf 
gocc-walle -o yt yt.bnf 
gocc-walle -a -o wproto wproto.bnf
