#!/usr/bin/env python
from abc import ABCMeta,abstractmethod
from arenasdk.exceptions.arena_exception import ArenaException
from arenasdk.enums.types import ArenaErrorType

class Field(object):
    __metaclass__ = ABCMeta  
    @abstractmethod
    def validate(self):
        pass  
      
    def options(self):
        pass  


class StringField(Field):
    def __init__(self,flag,value):
        super().__init__()
        self._flag = flag
        self._value = value
        
    def validate(self):
        if not self._flag or self._value == "":
            raise ArenaException(ArenaErrorType.ValidateArgsError,"failed to validate flag {},value is null".format(self._flag))
        return True
    def options(self):
        arena_options = list()
        arena_options.append(self._flag + "=" + self._value)
        return arena_options


class BoolField(Field):
    def __init__(self,flag):
        super().__init__()
        self._flag = flag

    def validate(self):
        return True 
    
    def options(self):
        arena_options = list()
        arena_options.append(self._flag)
        return arena_options

class StringListField(Field):
    def __init__(self,flag,values):
        super().__init__()
        self._flag = flag
        self._values = values
        
    def validate(self):
        if not self._values or len(self._values) == 0:
            raise ArenaException(ArenaErrorType,"failed to validate flag {},values are null".format(self._flag))
        return True
    
    def options(self):
        arena_options = list()
        for value in self._values:
            arena_options.append(self._flag + "=" + value)
        return arena_options

class StringMapField(Field):
    def __init__(self,flag,values,join_flag="="):
        super().__init__()
        self._flag = flag
        self._values = values
        self._join_flag = join_flag
    def validate(self):
        if not self._values or len(self._values) == 0:
            raise ArenaException(ArenaErrorType,"failed to validate flag {},values are null".format(self._flag))
        return True
    def options(self):
        arena_options = list()
        for key,value in self._values.items():
            arena_options.append(self._flag + "=" + key + self._join_flag + value)
        return arena_options 