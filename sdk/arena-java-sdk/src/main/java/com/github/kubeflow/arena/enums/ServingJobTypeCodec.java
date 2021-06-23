package com.github.kubeflow.arena.enums;

import com.alibaba.fastjson.parser.DefaultJSONParser;
import com.alibaba.fastjson.parser.deserializer.ObjectDeserializer;
import com.alibaba.fastjson.serializer.JSONSerializer;
import com.alibaba.fastjson.serializer.ObjectSerializer;

import java.io.IOException;
import java.lang.reflect.Type;

public class ServingJobTypeCodec implements ObjectSerializer, ObjectDeserializer {

    @Override
    public void write(JSONSerializer serializer, Object object, Object fieldName, Type fieldType, int features) throws IOException {
        serializer.write(((ServingJobType)object).alias());
    }

    @Override
    public <T> T deserialze(DefaultJSONParser parser, Type type, Object fieldName) {
        Object value = parser.parse();
        return value == null ? null : (T) ServingJobType.getByAlias((String) value);
    }

    @Override
    public int getFastMatchToken() {
        return 0;
    }
}
