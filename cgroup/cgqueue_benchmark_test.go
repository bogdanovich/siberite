package cgroup

import (
	"crypto/rand"
	"testing"
)

func Benchmark_CGQueue_Enqueue_1_Byte(b *testing.B) {
	q, _ := CGQueueOpen(cgQueueName, dir)
	defer q.Drop()
	value := make([]byte, 1)
	rand.Read(value)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Enqueue(value)
	}
}

func Benchmark_CGQueue_Enqueue_128_Bytes(b *testing.B) {
	q, _ := CGQueueOpen(cgQueueName, dir)
	defer q.Drop()
	value := make([]byte, 128)
	rand.Read(value)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Enqueue(value)
	}

}

func Benchmark_CGQueue_Enqueue_1024_Bytes(b *testing.B) {
	q, _ := CGQueueOpen(cgQueueName, dir)
	defer q.Drop()
	value := make([]byte, 1024)
	rand.Read(value)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Enqueue(value)
	}

}

func Benchmark_CGQueue_Enqueue_10240_Bytes(b *testing.B) {
	q, _ := CGQueueOpen(cgQueueName, dir)
	defer q.Drop()
	value := make([]byte, 10240)
	rand.Read(value)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Enqueue(value)
	}
}

func Benchmark_CGQueue_GetNext_1_Byte(b *testing.B) {
	q, _ := CGQueueOpen(cgQueueName, dir)
	defer q.Drop()
	value := make([]byte, 1)
	rand.Read(value)
	for i := 0; i < 500000; i++ {
		q.Enqueue(value)
	}

	q.Close()
	q, _ = CGQueueOpen(cgQueueName, dir)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.GetNext()
	}
}

func Benchmark_CGQueue_GetNext_128_Bytes(b *testing.B) {
	q, _ := CGQueueOpen(cgQueueName, dir)
	defer q.Drop()
	value := make([]byte, 128)
	rand.Read(value)
	for i := 0; i < 500000; i++ {
		q.Enqueue(value)
	}

	q.Close()
	q, _ = CGQueueOpen(cgQueueName, dir)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.GetNext()
	}
}

func Benchmark_CGQueue_GetNext_1024_Bytes(b *testing.B) {
	q, _ := CGQueueOpen(cgQueueName, dir)
	defer q.Drop()
	value := make([]byte, 1024)
	rand.Read(value)
	for i := 0; i < 200000; i++ {
		q.Enqueue(value)
	}

	q.Close()
	q, _ = CGQueueOpen(cgQueueName, dir)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.GetNext()
	}
}

func Benchmark_CGQueue_GetNext_10240_Bytes(b *testing.B) {
	q, _ := CGQueueOpen(cgQueueName, dir)
	defer q.Drop()
	value := make([]byte, 10240)
	rand.Read(value)
	for i := 0; i < 50000; i++ {
		q.Enqueue(value)
	}

	q.Close()
	q, _ = CGQueueOpen(cgQueueName, dir)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.GetNext()
	}
}

func Benchmark_CGQueue_ConsumerGroup_1_Byte(b *testing.B) {
	q, _ := CGQueueOpen(cgQueueName, dir)
	defer q.Drop()
	value := make([]byte, 1)
	rand.Read(value)
	for i := 0; i < 500000; i++ {
		q.Enqueue(value)
	}

	q.Close()
	q, _ = CGQueueOpen(cgQueueName, dir)
	cg, _ := q.ConsumerGroup("test")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cg.GetNext()
	}
}

func Benchmark_CGQueue_ConsumerGroup_128_Bytes(b *testing.B) {
	q, _ := CGQueueOpen(cgQueueName, dir)
	defer q.Drop()
	value := make([]byte, 128)
	rand.Read(value)
	for i := 0; i < 500000; i++ {
		q.Enqueue(value)
	}

	q.Close()
	q, _ = CGQueueOpen(cgQueueName, dir)
	cg, _ := q.ConsumerGroup("test")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cg.GetNext()
	}
}

func Benchmark_CGQueue_ConsumerGroup_1024_Bytes(b *testing.B) {
	q, _ := CGQueueOpen(cgQueueName, dir)
	defer q.Drop()
	value := make([]byte, 1024)
	rand.Read(value)
	for i := 0; i < 200000; i++ {
		q.Enqueue(value)
	}

	q.Close()
	q, _ = CGQueueOpen(cgQueueName, dir)
	cg, _ := q.ConsumerGroup("test")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cg.GetNext()
	}
}

func Benchmark_CGQueue_ConsumerGroup_10240_Bytes(b *testing.B) {
	q, _ := CGQueueOpen(cgQueueName, dir)
	defer q.Drop()
	value := make([]byte, 10240)
	rand.Read(value)
	for i := 0; i < 200000; i++ {
		q.Enqueue(value)
	}

	q.Close()
	q, _ = CGQueueOpen(cgQueueName, dir)
	cg, _ := q.ConsumerGroup("test")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cg.GetNext()
	}
}
