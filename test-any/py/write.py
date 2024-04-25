import schema_pb2

import google.protobuf.any_pb2 as any

person = schema_pb2.Person()
person.name = "John"
person.age = 42

any_person = any.Any()
any_person.Pack(person)

task1 = schema_pb2.Task(done_by=any_person)
task1.title = "task1"
task1.description = "task1 description"

with open("task1.bin", "wb") as f:
    f.write(task1.SerializeToString())
    print("task1.bin written")

robot = schema_pb2.Robot()
robot.name = "R2D2"

any_robot = any.Any()
any_robot.Pack(robot)

task2 = schema_pb2.Task(done_by=any_robot)
task2.title = "task2"
task2.description = "task2 description"

with open("task2.bin", "wb") as f:
    f.write(task2.SerializeToString())
    print("task2.bin written")