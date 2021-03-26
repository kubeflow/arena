#!/usr/bin/env python

class ArenaException(Exception):
    """ArenaException raised for errors when using arenasdk.

    Attributes:
        err_type -- the error type,it must be ArenaErrorType object
        err_msg -- the error message content
    """
    def __init__(self,err_type,err_msg):
        super().__init__(self)
        self.err_type = err_type
        self.err_msg = err_msg
    
    def __str__(self):
        return '''ArenaException[ErrorType: {},ErrorMessage: {}]'''.format(self.err_type.name,self.err_msg)
    
    def __repr__(self):
        return '''ArenaException[ErrorType: {},ErrorMessage: {}]'''.format(self.err_type.name,self.err_msg)
