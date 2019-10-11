/*******************************************************************************
 * Copyright 2019 Seldon Technologies Ltd (http://www.seldon.io/)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *******************************************************************************/
package io.seldon.engine.pb;

import com.google.protobuf.InvalidProtocolBufferException;

import io.kubernetes.client.proto.IntStr.IntOrString;

import org.junit.Assert;
import org.junit.Test;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.seldon.engine.pb.JsonFormat.Printer;

public class TestJsonFormat {
	private final static Logger logger = LoggerFactory.getLogger(TestJsonFormat.class);

	@Test
	public void testStrValCustomFormat() throws InvalidProtocolBufferException
	{
		final String val = "String Value";
		IntOrString is = IntOrString.newBuilder().setStrVal(val).build();
		Printer jf = JsonFormat.printer().usingTypeConverter(is.getDescriptorForType().getFullName(), new IntOrStringUtils.IntOrStringConverter());
		Assert.assertEquals("\""+val+"\"", jf.print(is));
	}

  @Test
  public void testEscapeHTML()  throws InvalidProtocolBufferException
  {
    final String val = "<div class=\"div-class\"></div>";
    final String escaped = "\\u003cdiv class\\u003d\\\"div-class\\\"\\u003e\\u003c/div\\u003e";
    final String expected = String.format("{\"strVal\":\"%s\"}", escaped);
		IntOrString is = IntOrString.newBuilder().setStrVal(val).build();
    Printer jf = JsonFormat.printer().omittingInsignificantWhitespace();
    final String json = jf.print(is);
    Assert.assertEquals(expected, json);
  }
	
	@Test
	public void testIntValCustomFormat() throws InvalidProtocolBufferException
	{
		final int val = 1;
		IntOrString is = IntOrString.newBuilder().setIntVal(val).build();
		Printer jf = JsonFormat.printer().usingTypeConverter(is.getDescriptorForType().getFullName(), new IntOrStringUtils.IntOrStringConverter());
		Assert.assertEquals(""+val, jf.print(is));
	}
	
	@Test
	public void testIntValDefaultFormat() throws InvalidProtocolBufferException
	{
		final int val = 1;
		IntOrString is = IntOrString.newBuilder().setIntVal(val).build();
		Printer jf = JsonFormat.printer().omittingInsignificantWhitespace();
		Assert.assertEquals("{\"intVal\":"+val+"}", jf.print(is));
	}
	
	@Test
	public void testStrValDefaultFormat() throws InvalidProtocolBufferException
	{
		final String val = "String Value";
		IntOrString is = IntOrString.newBuilder().setStrVal(val).build();
		Printer jf = JsonFormat.printer().omittingInsignificantWhitespace();
		Assert.assertEquals("{\"strVal\":\""+val+"\"}", jf.print(is));
	}
}

