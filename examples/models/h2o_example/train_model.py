"""This script run the code in https://github.com/h2oai/h2o-tutorials/blob/master/h2o-open-tour-2016/chicago/intro-to-h2o.ipynb
   and save the trained model in a file glm_fit1 in the same directory of the script.

   Data is not split into train and test sets as it is irrelevant for the purpose of this example.
   Instead, training is performed on the whole dataset.
"""

# Load the H2O library and start up the H2O cluter locally on your machine
import h2o
# Import H2O GLM:
from h2o.estimators.glm import H2OGeneralizedLinearEstimator

if __name__=="__main__":

    # Number of threads, nthreads = -1, means use all cores on your machine
    # max_mem_size is the maximum memory (in GB) to allocate to H2O
    h2o.init(nthreads = -1, max_mem_size = 8)
    
    #loan_csv = "/Volumes/H2OTOUR/loan.csv"  # modify this for your machine
    # Alternatively, you can import the data directly from a URL
    loan_csv = "https://raw.githubusercontent.com/h2oai/app-consumer-loan/master/data/loan.csv"
    data = h2o.import_file(loan_csv)  # 163,987 rows x 15 columns
    data['bad_loan'] = data['bad_loan'].asfactor()  #encode the binary response as a factor
    #data['bad_loan'].levels()  #optional: after encoding, this shows the two factor levels, '0' and '1'
    
    y = 'bad_loan'
    x = list(data.columns)
    x.remove(y)  #remove the response
    x.remove('int_rate')  #remove the interest rate column because it's correlated with the outcome
    
    # Initialize the GLM estimator:
    # Similar to R's glm() and H2O's R GLM, H2O's GLM has the "family" argument
    glm_fit1 = H2OGeneralizedLinearEstimator(family='binomial', model_id='glm_fit1')
    glm_fit1.train(x=x, y=y, training_frame=data)
    
    model_path = h2o.save_model(model=glm_fit1, path="", force=True)
