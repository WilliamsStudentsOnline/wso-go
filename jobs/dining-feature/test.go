package main

import (
  "fmt"
)

type Person struct {
  firstName string
  lastName string
  uniqueID int
}

func getPerson(id int, addresses []*Person) *Person {
  var nilGuy *Person
  for _, person := range addresses{
    if person == nilGuy {
      fmt.Printf("Found a nil one! Addy: %p\n", nilGuy)
      break
    }
    if (*person).uniqueID == 23{
      fmt.Printf("Ammar's memory address: %p\n", person)
      return person
    }
  }
  fmt.Printf("Returning nil! \n", nilGuy)
  return nilGuy
}

func main() {
    a := Person{ firstName: "Ammar", lastName: "Eltigani", uniqueID: 23,}
    fmt.Printf("Ammar's memory address: %p\n", &a)

    addresses := make([]*Person, 100)
    addresses[0] =  &a

    var personFound *Person
    personFound = getPerson(23, addresses)
    fmt.Printf("personFound's (Ammar) memory address: %p\n", personFound)

    (*personFound).firstName = "Nathan"
    fmt.Printf("addresses[0]: %+v", (*addresses[0]))
  }



  // primes := [6]int{2, 3, 5, 7, 11, 13}
	// var s []int = primes[1:4]
  //
  // // Slices have dynamic size. Arrays and slices each have advantages
  //    // but use cases for slices are much more common.
  //    s3 := []int{4, 5, 9}    // Compare to a5. No ellipsis here.
  //    s4 := make([]int, 4)    // Allocates slice of 4 ints, initialized to all 0.
  //    var d2 [][]float64      // Declaration only, nothing allocated here.
  //    bs := []byte("a slice") // Type conversion syntax.
