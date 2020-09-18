CGO_ENABLED=0 \
GOOS=linux \
GOARCH=amd64 \
go build -mod=readonly \
-ldflags "-X github.com/KuChainNetwork/kuchain/chain/constants/keys.ChainNameStr=kts -X github.com/KuChainNetwork/kuchain/chain/constants/keys.ChainMainNameStr=kratos" \
-o kucd github.com/KuChainNetwork/kuchain/cmd/kucd && \
upx kucd && \
scp kucd root@ks02:/data/db_backend && \
ssh -t root@ks02 "cd /data/db_backend && ./bi.sh && docker-compose down && docker-compose up -d"