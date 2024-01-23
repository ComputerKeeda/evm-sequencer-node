sudo rm -rf data/leveldb
mkdir -p data/leveldb/batches
mkdir -p data/leveldb/blocks
mkdir -p data/leveldb/proof
mkdir -p data/leveldb/publicWitness
mkdir -p data/leveldb/static
mkdir -p data/leveldb/tx
touch data/blockCount.txt
touch data/transactionCount.txt
touch data/batchCount.txt
echo "0" > data/blockCount.txt
echo "0" > data/transactionCount.txt
echo "0" > data/batchCount.txt
go run main.go