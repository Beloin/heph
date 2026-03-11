
# TODO: bsena; Add transitive when you know how to create maps/structs
def target(
    name: str,
    driver: str,
    labels: list[str],
    *args,
    **kwargs
):
    """Define a target for execution in the heph build system.

    A target is defined by a name, and a driver. The driver specification dictates other args and commands.
    This execution unit is isolated from the rest of the repo which allows for efficient caching and parallel execution.

    Args:
        name (str): Target name (required)
        driver (str): The driver to execute this target (required)
        labels (list[str]): A list of labels to attach to this target (required)
        *args (any): Arguments to be sent directly into the driver when executed.
        **kwargs (any): Keyword arguments to be passed directly to driver.
    """
    pass

