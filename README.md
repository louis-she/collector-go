Log collector && Analysor
=========================

What is this
------------

This is a software used as log collector and analysor, it constantly "tail" some specified files and then apply the content to a chain of content handler. The handler and file name parser are totally free to add. Users just focus on what should do with a line of the log and maybe how to solve the dynamic file name.


How to use
----------

The config file is `conf.json` and it looks like this,

```json
{
  "Entity": [
    {   
      "Path": ["/tmp/file_A", "/tmp/%Y%zM%zD_file_B"],
      "PathParser": ["Timefmt"],
      "HandlerChain": [
        ["some_handler_1", "some_handler_2()"],
        ["some_handler_1", "some_handler_4"],
      ],  
      "Timespan": 2
    },  
    {   
      "Path": ["/tmp/file_C"],
      "Empty": "OuputStreamNotRunning",
      "Timespan": 4
    }   
  ]
}
```

* Entity 

This stands for the collector entity, just put it there.

* PathParser[Optional]

This is the file name parser, will be useful when the file name is dynamic. Say that you have a log file which the file name is based on the date. Like **/tmp/20160112_file_B**, then you should add a PathParser for it as in the above sample conf.json, there is already a file parser for the date **Timefmt** , you can use it directly or write your own. If you need to add multiple parser for a file, you can chain the parser together.

* HandlerChain

A handler is a func which take just one parameter of string type, which is a line of the log. You can do whatever you want with this line. And the return value should also be a string type.

  * **Handlers can be chained** In the above example, the **some_handler_4** will be executed right after **some_handler_1**, the parameter of **some_handler_4** will be the the return value of **some_handler_1**
  * **Handlers can be higher order function** You may notice that there is a **()** in the **some_handler_2()**, the **()** makes the **some_handler_2** a higher order function, then the **some_handler_2** should not return a string but a function, and this return function is actually the real handler. This can be helpful when you need some variables be accessed for a handler in the whole time the collector is running, without being the global variables. Say that you want to sum the lines you have already "tailed", then you should put a **sum** variable in the super order function and access the variable in the return function which is the handler.
  * **Result cached** In the example, there are two handler chain for one file, and for each of the chain, the first handler is **some_handler_1**. In this case, there is no need to execute the **some_handler_1** twice, and the collector do will cache the result of **some_handler_1**

* Timespan

The span of a entity execution.

* Empty

This is actually a special handler, that is when there is no content received while tail(there was no content added to the log from the last execution till this time), then the Empty handler will be executed.
