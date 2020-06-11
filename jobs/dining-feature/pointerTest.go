package main

import (
  "fmt"
  "reflect"
)

type People struct {
  people []*Person
}

type Person struct {
  firstName string
  lastName string
  uniqueID int
}

func getPerson(id int, personArr []*Person) *Person {
  var nilGuy *Person
  for _,existingPerson := range personArr {
    if existingPerson == nilGuy {
      fmt.Printf("Found a nil one! Addy: %p\n", nilGuy)
      break
    }
    if (*existingPerson).uniqueID == id {
       fmt.Printf("I found %s with the address: %p\n\n", existingPerson.firstName, existingPerson)
       return existingPerson
    }
  }
  fmt.Printf("failed getting a person with the id %d: ", id)
  var nonsense *Person
  return nonsense
}

func main() {
  var students People

  nate := Person{ firstName: "Nathan", lastName: "Thimothe", uniqueID: 666,}
  students.people = make([]*Person, 100)
  students.people[0] = &nate

  var nateTwo *Person = new(Person)
  (*nateTwo).firstName = "woah"
  // THE FOLLOWING WORKS: NOT SURE WHY
  // var nateTwo Person
  // nateTwo.firstName = "woah"
  // nateTwo.lastName = "there"
  // (&nateTwo).firstName = "HEY"
  fmt.Printf("natTwo: %+v", nateTwo)
  // OR students.people = append(students.people, &nate)
  fmt.Println(reflect.ValueOf(students.people).Kind())
  //fmt.Println(reflect.ValueOf(slice).Kind())
  //fmt.Println(slice)
  // fmt.Printf("%s's memory address: %p\n", nate.firstName, &nate)
  // fmt.Printf("Students.people memory address: %p\n", &students.people)
  // students.people = append(students.people, &nate)
  //fmt.Printf("Students.people memory address 2: %p\n", &students.people)

  var personFound *Person = new(Person)
  fmt.Printf("personFound's before getPerson memory address: %p\n", personFound)
  personFound = getPerson(666, students.people)
  fmt.Printf("personFound's (Nathan) memory address: %p\n", personFound)
}


  // fmt.Printf("People's memory addy: %p | People.people's memory addy: %p | People.people's contents : %+v\n\n",&people, &people.people, people.people)
  // people.people = append(people.people, Person{firstName : "BahOuais", lastName: "", uniqueID: 4,})
  //
  // var personA Person
  // personA.uniqueID = 4
  //
  // var personB *Person
  // fmt.Printf("This is PersonB's memory address: %p \n\n\n", personB)
  //
  // personExists := personExists(personA, people.people)
  // // personA exists in people.people
  // if !personExists {
  //   personA.firstName = "Nice"
  //   people.people = append(people.people, personA)
  //   } else {
  //     personB = getPerson(personA, people.people)
  //   }
  //   //personA.lastName = "maisNon"
  //
  //   personB.lastName = "J'ai réussi"
  //
  //   fmt.Printf("This is PersonB's memory address: %p \nThis is PersonB: %+v\n\n", personB,*personB)
  //   fmt.Printf("People.people's memory addy: %p \nPeople.people's contents : %+v\n",&people.people, people.people)
