package main

import (
  "testing"
)

func BenchmarkDecode(b *testing.B) {
  for i := 0; i < b.N; i++ {
    decode("post.json") 
  }
}

func BenchmarkUnmarshal(b *testing.B) {
  for i := 0; i < b.N; i++ {
    unmarshal("post.json")
  }
}

func BenchmarkFibinacchiIterative(b *testing.B) {
  for i := 0; i < b.N; i++ {
    fibonacciIterative(20)
  }
}

func BenchmarkFibinacchiRecursive(b *testing.B) {
  for i := 0; i < b.N; i++ {
    fibonacciRecursive(20)
  }
}
