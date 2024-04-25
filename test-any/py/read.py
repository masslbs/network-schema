import schema_pb2

def decode(fname):
  data = None
  with open(fname, "rb") as f:
    data = f.read()

  assert data is not None
  task1 = schema_pb2.Task()
  task1.ParseFromString(data)
  print(task1)
  typeName = task1.done_by.TypeName()
  print("\nby type:",typeName)

  print("\ndone_by:")
  if typeName == "tutorial.Person":
    person = schema_pb2.Person()
    task1.done_by.Unpack(person)
    print(person)
  elif typeName == "tutorial.Robot":
    robot = schema_pb2.Robot()
    task1.done_by.Unpack(robot)
    print(robot)
  else:
    print(f"unknown type {typeName}")

decode("task1.bin")
decode("task2.bin")
