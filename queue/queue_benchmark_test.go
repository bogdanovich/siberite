package queue

import (
	"crypto/rand"
	"testing"
)

func Benchmark_Queue_Enqueue_1_Byte(b *testing.B) {
	q, _ := Open(name, dir, &options)
	defer q.Drop()
	value := make([]byte, 1)
	rand.Read(value)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Enqueue(value)
	}
}

func Benchmark_Queue_Enqueue_128_Bytes(b *testing.B) {
	q, _ := Open(name, dir, &options)
	defer q.Drop()
	value := make([]byte, 128)
	rand.Read(value)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Enqueue(value)
	}

}

func Benchmark_Queue_Enqueue_1024_Bytes(b *testing.B) {
	q, _ := Open(name, dir, &options)
	defer q.Drop()
	value := make([]byte, 1024)
	rand.Read(value)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Enqueue(value)
	}

}

func Benchmark_Queue_Enqueue_10240_Bytes(b *testing.B) {
	q, _ := Open(name, dir, &options)
	defer q.Drop()
	value := make([]byte, 10240)
	rand.Read(value)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Enqueue(value)
	}
}

func Benchmark_Queue_Enqueue_102400_Bytes(b *testing.B) {
	q, _ := Open(name, dir, &options)
	defer q.Drop()
	value := make([]byte, 102400)
	rand.Read(value)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Enqueue(value)
	}
}

func Benchmark_Queue_GetNext_1_Byte(b *testing.B) {
	q, _ := Open(name, dir, &options)
	defer q.Drop()
	value := make([]byte, 1)
	rand.Read(value)
	for i := 0; i < 500000; i++ {
		q.Enqueue(value)
	}

	q.Close()
	q, _ = Open(name, dir, &options)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.GetNext()
	}
}

func Benchmark_Queue_GetNext_128_Bytes(b *testing.B) {
	q, _ := Open(name, dir, &options)
	defer q.Drop()
	value := make([]byte, 128)
	rand.Read(value)
	for i := 0; i < 500000; i++ {
		q.Enqueue(value)
	}

	q.Close()
	q, _ = Open(name, dir, &options)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.GetNext()
	}
}

func Benchmark_Queue_GetNext_1024_Bytes(b *testing.B) {
	q, _ := Open(name, dir, &options)
	defer q.Drop()
	value := make([]byte, 1024)
	rand.Read(value)
	for i := 0; i < 200000; i++ {
		q.Enqueue(value)
	}

	q.Close()
	q, _ = Open(name, dir, &options)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.GetNext()
	}
}

func Benchmark_Queue_GetNext_10240_Bytes(b *testing.B) {
	q, _ := Open(name, dir, &options)
	defer q.Drop()
	value := make([]byte, 10240)
	rand.Read(value)
	for i := 0; i < 200000; i++ {
		q.Enqueue(value)
	}

	q.Close()
	q, _ = Open(name, dir, &options)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.GetNext()
	}
}
