library(rpart)

data(iris)
names(iris) <- tolower(sub('.', '_', names(iris), fixed = TRUE))
fit <- rpart(species ~ ., iris)
saveRDS(fit, file = "model.Rds", compress = TRUE)
