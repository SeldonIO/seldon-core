/**
 * ***************************************************************************** Copyright 2017
 * Seldon Technologies Ltd (http://www.seldon.io/)
 *
 * <p>Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
 * except in compliance with the License. You may obtain a copy of the License at
 *
 * <p>http://www.apache.org/licenses/LICENSE-2.0
 *
 * <p>Unless required by applicable law or agreed to in writing, software distributed under the
 * License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
 * express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 * *****************************************************************************
 */
package io.seldon.engine.predictors;

import com.google.protobuf.ListValue;
import com.google.protobuf.Value;
import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.DefaultData.DataOneofCase;
import io.seldon.protos.PredictionProtos.Tensor;
import java.util.Iterator;
import java.util.List;
import org.ojalgo.matrix.Primitive64Matrix;

public class PredictorUtils {

  public static DefaultData nDArrayToTensor(DefaultData data) {
    if (data.getDataOneofCase() == DataOneofCase.TENSOR) {
      return data;

    } else if (data.getDataOneofCase() == DataOneofCase.NDARRAY) {
      int bLength = data.getNdarray().getValuesCount();
      int vLength = data.getNdarray().getValues(0).getListValue().getValuesCount();

      DefaultData.Builder dataBuilder = DefaultData.newBuilder();
      Tensor.Builder tBuilder = Tensor.newBuilder().addShape(bLength).addShape(vLength);

      int index = 0;
      for (Iterator<String> i = data.getNamesList().iterator(); i.hasNext(); ) {
        dataBuilder.setNames(index, i.next());
        index++;
      }

      for (int i = 0; i < bLength; ++i) {
        for (int j = 0; j < vLength; ++j) {
          tBuilder.addValues(
              data.getNdarray().getValues(i).getListValue().getValues(j).getNumberValue());
        }
      }

      dataBuilder.setTensor(tBuilder);

      return dataBuilder.build();
    }
    return null;
  }

  public static void add(DefaultData data, Primitive64Matrix.DenseReceiver receiver) {

    if (data.getDataOneofCase() == DataOneofCase.TENSOR) {

      List<Double> valuesList = data.getTensor().getValuesList();
      List<Integer> shapeList = data.getTensor().getShapeList();

      int rows = shapeList.get(0);
      int cols = shapeList.get(1);

      for (int i = 0; i < rows * cols; i++) {
        receiver.add(i / cols, i % cols, valuesList.get(i));
      }

    } else if (data.getDataOneofCase() == DataOneofCase.NDARRAY) {

      ListValue list = data.getNdarray();

      int rows = list.getValuesCount();
      int cols = list.getValues(0).getListValue().getValuesCount();

      for (int i = 0; i < rows; ++i) {
        ListValue rowListValue = list.getValues(i).getListValue();
        for (int j = 0; j < cols; j++) {
          receiver.add(i, j, rowListValue.getValues(j).getNumberValue());
        }
      }
    }
  }

  public static Primitive64Matrix getOJMatrix(DefaultData data) {
    Primitive64Matrix.Factory matrixFactory = Primitive64Matrix.FACTORY;
    if (data.getDataOneofCase() == DataOneofCase.TENSOR) {

      List<Double> valuesList = data.getTensor().getValuesList();
      List<Integer> shapeList = data.getTensor().getShapeList();

      int rows = shapeList.get(0);
      int columns = shapeList.get(1);

      double[][] values = new double[rows][columns];
      for (int i = 0; i < rows * columns; i++) {
        values[i / columns][i % columns] = valuesList.get(i);
      }

      return matrixFactory.rows(values);
    } else if (data.getDataOneofCase() == DataOneofCase.NDARRAY) {
      ListValue list = data.getNdarray();
      int rows = list.getValuesCount();
      int cols = list.getValues(0).getListValue().getValuesCount();

      double[][] values = new double[rows][cols];
      for (int i = 0; i < rows; ++i) {
        for (int j = 0; j < cols; j++) {
          values[i][j] = list.getValues(i).getListValue().getValues(j).getNumberValue();
        }
      }

      return matrixFactory.rows(values);
    }
    return null;
  }

  public static int[] getShape(DefaultData data) {
    if (data.getDataOneofCase() == DataOneofCase.TENSOR) {
      List<Integer> shapeList = data.getTensor().getShapeList();
      int[] shape = new int[shapeList.size()];
      for (int i = 0; i < shape.length; ++i) {
        shape[i] = shapeList.get(i);
      }

      return shape;
    } else if (data.getDataOneofCase() == DataOneofCase.NDARRAY) {
      int bLength = data.getNdarray().getValuesCount();
      int vLength = data.getNdarray().getValues(0).getListValue().getValuesCount();

      int[] shape = {bLength, vLength};
      return shape;
    }
    return null;
  }

  public static DefaultData updateData(DefaultData oldData, Primitive64Matrix newData) {
    DefaultData.Builder dataBuilder = DefaultData.newBuilder();

    dataBuilder.addAllNames(oldData.getNamesList());

    //		int index=0;
    //		for (Iterator<String> i = oldData.getFeaturesList().iterator(); i.hasNext();){
    //			dataBuilder.setFeatures(index, i.next());
    //			index++;
    //		}

    int rows = (int) newData.countRows();
    int cols = (int) newData.countColumns();

    if (oldData.getDataOneofCase() == DataOneofCase.TENSOR) {
      Tensor.Builder tBuilder = Tensor.newBuilder();

      tBuilder.addShape(rows);
      tBuilder.addShape(cols);

      for (int i = 0; i < rows; ++i) {
        for (int j = 0; j < cols; ++j) {
          tBuilder.addValues(newData.doubleValue(i, j));
        }
      }
      dataBuilder.setTensor(tBuilder);
      return dataBuilder.build();
    } else if (oldData.getDataOneofCase() == DataOneofCase.NDARRAY) {
      ListValue.Builder b1 = ListValue.newBuilder();
      for (int i = 0; i < rows; ++i) {
        ListValue.Builder b2 = ListValue.newBuilder();
        for (int j = 0; j < cols; j++) {
          b2.addValues(Value.newBuilder().setNumberValue(newData.doubleValue(i, j)));
        }
        b1.addValues(Value.newBuilder().setListValue(b2.build()));
      }
      dataBuilder.setNdarray(b1.build());
      return dataBuilder.build();
    }
    return null;
  }
}
