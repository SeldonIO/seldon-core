package io.seldon.wrapper.utils;

import java.util.ArrayList;
import java.util.List;

import org.nd4j.linalg.dataset.api.iterator.CachingDataSetIterator;

import com.google.protobuf.ListValue;
import com.google.protobuf.Value;

import hex.genmodel.easy.RowData;
import hex.genmodel.easy.prediction.AbstractPrediction;
import hex.genmodel.easy.prediction.BinomialModelPrediction;
import hex.genmodel.easy.prediction.ClusteringModelPrediction;
import hex.genmodel.easy.prediction.DimReductionModelPrediction;
import hex.genmodel.easy.prediction.MultinomialModelPrediction;
import hex.genmodel.easy.prediction.OrdinalModelPrediction;
import hex.genmodel.easy.prediction.RegressionModelPrediction;
import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.Tensor;
import io.seldon.protos.PredictionProtos.DefaultData.DataOneofCase;

/**
 * Utilities for working with H2O models
 * 
 * @author clive
 *
 */
public class H2OUtils {
	
	/**
	 * Convert a Seldon Default Data into H2O RowData
	 * @param data Seldon protobuf data
	 * @return List of H2O RowData
	 */
	public static List<RowData> convertSeldonMessage(DefaultData data) {
		List<RowData> out = new ArrayList<>();
		if (data.getDataOneofCase() == DataOneofCase.TENSOR) {

			List<Double> valuesList = data.getTensor().getValuesList();
			List<Integer> shapeList = data.getTensor().getShapeList();

			if (shapeList.size() > 2) {
				return null;
			}

			int cols = 0;
			if (shapeList.size() == 1)
				cols = shapeList.get(0);
			else
				cols = shapeList.get(1);
			RowData row = new RowData();
			for (int i = 0; i < valuesList.size(); i++) {
				if (i > 0 && i % cols == 0) {
					out.add(row);
					row = new RowData();
				}
				String name = data.getNamesCount() > 0 ? data.getNames(i % cols) : "" + (i % cols);
				row.put(name, valuesList.get(i));
			}
			out.add(row);

			return out;
		} else if (data.getDataOneofCase() == DataOneofCase.NDARRAY) {
			ListValue list = data.getNdarray();
			int rows = list.getValuesCount();
			int cols = list.getValues(0).getListValue().getValuesCount();

			for (int i = 0; i < rows; ++i) {
				RowData row = new RowData();
				for (int j = 0; j < cols; j++) {
					String name = data.getNamesCount() > 0 ? data.getNames(j % cols) : "" + (j % cols);
					Object value;
					Value listValue = list.getValues(i).getListValue().getValues(j);
					switch(listValue.getKindCase()) {
					case NUMBER_VALUE:
						value = listValue.getNumberValue();
						break;
					case STRING_VALUE:
						value = listValue.getStringValue();
						break;
					case BOOL_VALUE:
						//Get value as String
						value = listValue.getStringValue();
						break;
					case NULL_VALUE:
						// Treat Nulls as 0
						value = 0.0;
						break;
					case LIST_VALUE:
						throw new UnsupportedOperationException("Only 2-D arrays unsupported for H2O conversion");
					case STRUCT_VALUE:
						throw new UnsupportedOperationException("Struct in NDArray unsupported for H2O conversion");
					default:
						throw new UnsupportedOperationException("Unknown kind in NDArray");
					}
					row.put(name, value);
				}
				out.add(row);
			}
			return out;
		} else
			return null;
	}

	/**
	 * Convert a prediction result from H2O to Seldon protobuf DefaultData with same type as input
	 * @param predictions The H2O predictions
	 * @param input The original input
	 * @return A seldon DefaultData protobuf message
	 */
	public static DefaultData convertH2OPrediction(List<AbstractPrediction> predictions, DefaultData input) {
		if (input == null || input.getDataOneofCase() == DataOneofCase.TENSOR) {
			int rows = predictions.size();
			Tensor.Builder tBuilder = Tensor.newBuilder();
			for (AbstractPrediction p : predictions) {
				if (p instanceof BinomialModelPrediction) {
					BinomialModelPrediction bp = (BinomialModelPrediction) p;
					for (int i = 0; i < bp.classProbabilities.length; i++)
						tBuilder.addValues(bp.classProbabilities[i]);
				} else if (p instanceof MultinomialModelPrediction) {
					MultinomialModelPrediction mp = (MultinomialModelPrediction) p;
					for (int i = 0; i < mp.classProbabilities.length; i++)
						tBuilder.addValues(mp.classProbabilities[i]);
				} else if (p instanceof OrdinalModelPrediction) {
					OrdinalModelPrediction op = (OrdinalModelPrediction) p;
					for (int i = 0; i < op.classProbabilities.length; i++)
						tBuilder.addValues(op.classProbabilities[i]);
				} else if (p instanceof ClusteringModelPrediction) {
					ClusteringModelPrediction cp = (ClusteringModelPrediction) p;
					for (int i = 0; i < cp.distances.length; i++)
						tBuilder.addValues(cp.distances[i]);
				} else if (p instanceof RegressionModelPrediction) {
					RegressionModelPrediction r = (RegressionModelPrediction) p;
					tBuilder.addValues(r.value);
				} else if (p instanceof DimReductionModelPrediction) {
					DimReductionModelPrediction cp = (DimReductionModelPrediction) p;
					for (int i = 0; i < cp.dimensions.length; i++)
						tBuilder.addValues(cp.dimensions[i]);
				} else
					return null;
			}
			DefaultData.Builder dataBuilder = DefaultData.newBuilder();
			dataBuilder.setTensor(tBuilder);
			return dataBuilder.build();
		} else if (input.getDataOneofCase() == DataOneofCase.NDARRAY) {
			ListValue.Builder rows = ListValue.newBuilder();
			for (AbstractPrediction p : predictions) {
				ListValue.Builder row = ListValue.newBuilder();
				if (p instanceof BinomialModelPrediction) {
					BinomialModelPrediction bp = (BinomialModelPrediction) p;
					for (int i = 0; i < bp.classProbabilities.length; i++)
						row.addValues(Value.newBuilder().setNumberValue(bp.classProbabilities[i]));
				} else if (p instanceof MultinomialModelPrediction) {
					MultinomialModelPrediction mp = (MultinomialModelPrediction) p;
					for (int i = 0; i < mp.classProbabilities.length; i++)
						row.addValues(Value.newBuilder().setNumberValue(mp.classProbabilities[i]));
				} else if (p instanceof OrdinalModelPrediction) {
					OrdinalModelPrediction op = (OrdinalModelPrediction) p;
					for (int i = 0; i < op.classProbabilities.length; i++)
						row.addValues(Value.newBuilder().setNumberValue(op.classProbabilities[i]));
				} else if (p instanceof ClusteringModelPrediction) {
					ClusteringModelPrediction cp = (ClusteringModelPrediction) p;
					for (int i = 0; i < cp.distances.length; i++)
						row.addValues(Value.newBuilder().setNumberValue(cp.distances[i]));
				} else if (p instanceof RegressionModelPrediction) {
					RegressionModelPrediction r = (RegressionModelPrediction) p;
					row.addValues(Value.newBuilder().setNumberValue(r.value));
				} else if (p instanceof DimReductionModelPrediction) {
					DimReductionModelPrediction cp = (DimReductionModelPrediction) p;
					for (int i = 0; i < cp.dimensions.length; i++)
						row.addValues(Value.newBuilder().setNumberValue(cp.dimensions[i]));
				} else
					return null;
				rows.addValues(Value.newBuilder().setListValue(row.build()));
			}
			DefaultData.Builder dataBuilder = DefaultData.newBuilder();
			dataBuilder.setNdarray(rows.build());
			return dataBuilder.build();
		} else
			return null;
	}
}
