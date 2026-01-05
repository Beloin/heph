import types

def my_custom_function(arg1: str, arg2) -> str:
    """
    My custom comment
    """
    return "wrapped: " + arg1 + arg2


def my_other_function(arg1, arg2, arg3=12, arg4: str = "abc"):
    pass


def my_argless_function():
    pass


my_custom_variable = "custom variable value"

# My Variable Comment
my_new_var = 12

my_custom_result = my_custom_function(my_custom_variable, "")

print(my_custom_result)

