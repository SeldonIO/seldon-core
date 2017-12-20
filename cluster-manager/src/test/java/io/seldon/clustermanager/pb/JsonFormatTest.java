/*******************************************************************************
 * Copyright 2017 Seldon Technologies Ltd (http://www.seldon.io/)
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
package io.seldon.clustermanager.pb;

import org.junit.Assert;
import org.junit.Test;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.google.protobuf.InvalidProtocolBufferException;

import io.kubernetes.client.proto.IntStr.IntOrString;
import io.kubernetes.client.proto.Resource.Quantity;
import io.seldon.clustermanager.pb.JsonFormat.Printer;

public class JsonFormatTest {
	private final static Logger logger = LoggerFactory.getLogger(JsonFormatTest.class);
	
	@Test
	public void testStrValCustomFormat() throws InvalidProtocolBufferException
	{
		final String val = "String Value";
		IntOrString is = IntOrString.newBuilder().setStrVal(val).build();
		Printer jf = JsonFormat.printer().usingTypeConverter(is.getDescriptorForType().getFullName(), new IntOrStringUtils.IntOrStringConverter());
		Assert.assertTrue(jf.print(is).equals("\""+val+"\""));
	}
	
	@Test
	public void testIntValCustomFormat() throws InvalidProtocolBufferException
	{
		final int val = 1;
		IntOrString is = IntOrString.newBuilder().setIntVal(val).build();
		Printer jf = JsonFormat.printer().usingTypeConverter(is.getDescriptorForType().getFullName(), new IntOrStringUtils.IntOrStringConverter());
		Assert.assertTrue(jf.print(is).equals(""+val));
	}
	
	@Test
	public void testIntValDefaultFormat() throws InvalidProtocolBufferException
	{
		final int val = 1;
		IntOrString is = IntOrString.newBuilder().setIntVal(val).build();
		Printer jf = JsonFormat.printer().omittingInsignificantWhitespace();
		Assert.assertTrue(jf.print(is).equals("{\"intVal\":"+val+"}"));
	}
	
	@Test
	public void testStrValDefaultFormat() throws InvalidProtocolBufferException
	{
		final String val = "String Value";
		IntOrString is = IntOrString.newBuilder().setStrVal(val).build();
		Printer jf = JsonFormat.printer().omittingInsignificantWhitespace();
		Assert.assertTrue(jf.print(is).equals("{\"strVal\":\""+val+"\"}"));
	}
	
	@Test
	public void testQuantityCustomFormat() throws InvalidProtocolBufferException
	{
		final String val = "100Mi";
		Quantity q = Quantity.newBuilder().setString(val).build();
		Printer jf = JsonFormat.printer().usingTypeConverter(q.getDescriptorForType().getFullName(), new QuantityUtils.QuantityConverter());
		Assert.assertTrue(jf.print(q).equals("\""+val+"\""));
	}
	
	@Test
	public void testIntStrParseForString() throws InvalidProtocolBufferException
	{
		String val = "string Value";
		String json = "\""+val+"\"";
		IntOrString.Builder builder = IntOrString.newBuilder();
		JsonFormat.parser().usingTypeParser(IntOrString.getDescriptor().getFullName(),new IntOrStringUtils.IntOrStringParser()).merge(json, builder);
		Assert.assertEquals(val, builder.build().getStrVal());
	}
	
	@Test
	public void testIntStrParseForInt() throws InvalidProtocolBufferException
	{
		int val = 42;
		String json = ""+val;
		IntOrString.Builder builder = IntOrString.newBuilder();
		JsonFormat.parser().usingTypeParser(IntOrString.getDescriptor().getFullName(),new IntOrStringUtils.IntOrStringParser()).merge(json, builder);
		Assert.assertEquals(val, builder.build().getIntVal());
	}
	
	@Test
	public void testQuantityParseForString() throws InvalidProtocolBufferException
	{
		String val = "100Mi";
		String json = "\""+val+"\"";
		Quantity.Builder builder = Quantity.newBuilder();
		JsonFormat.parser().usingTypeParser(Quantity.getDescriptor().getFullName(),new QuantityUtils.QuantityParser()).merge(json, builder);
		Assert.assertEquals(val, builder.build().getString());
	}
	
}
