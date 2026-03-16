load("//folder/to/load", "my_func")

print("My Load")

my_other_call(arg1, kwarg1="literal", kwarg2=12)

def my_func(param1: str, param2: Union[str, int], param3: List[str] = []):
    inner_call(param1, kwarg="value")
    another_inner_call()

fun_no_args()

a = target(arg1)

b = fun_with_no_args()

c = fun_with_no_args()
