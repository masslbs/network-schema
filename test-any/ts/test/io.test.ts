import { assert, test } from "vitest";
import { tutorial } from '../src/compiled'

import * as fs from 'node:fs/promises';


test("encode", () => {
  const p = new tutorial.Person();
  p.name = "john"
  p.age = 20
  p.id = 23
  // console.log(p)

  const t = new tutorial.Task();
  t.title = "task1"
  t.doneBy = {
    type_url: "type.googleapis.com/tutorial.Person",
    value: tutorial.Person.encode(p).finish(),
  }
  // console.log(t)
  
  const bytes = tutorial.Task.encode(t).finish()
  // console.log(bytes.toString("base64"))

  const t2 = tutorial.Task.decode(bytes)
  // console.log(t2)
  assert(t2.title === t.title)
  assert(t2.doneBy!.type_url === t.doneBy!.type_url)
})

test("read frm file", async () => {
  const data = await fs.readFile("../py/task1.bin")
  console.log("read data:")
  console.log(data)

  const t = tutorial.Task.decode(data)
  assert(t.title === "task1")
  assert(t.doneBy!.type_url === "type.googleapis.com/tutorial.Person")
  console.log(t)

  let p = tutorial.Person.decode(t.doneBy!.value!)
  console.log(p)
  assert(p.name === "John")
})
