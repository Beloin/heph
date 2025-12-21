import types

def my_custom_function(arg1: str, arg2) -> str:
    """
    My custom comment
    """
    return "wrapped: " + arg1 + arg2


my_custom_variable = "custom variable value"


my_custom_result = my_custom_function(my_custom_variable, "")

print(my_custom_result)

