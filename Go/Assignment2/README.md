# Usage
Open 6 terminal and in each run one of this lines

p1 - go run peer.go localhost:8081 localhost:8082   
p2 - go run peer.go localhost:8082 localhost:8081 localhost:8083 localhost:8084
p3 - go run peer.go localhost:8083 localhost:8082   
p4 - go run peer.go localhost:8084 localhost:8082 localhost:8085 localhost:8086
p5 - go run peer.go localhost:8085 localhost:8084    
p6 - go run peer.go localhost:8086 localhost:8084  