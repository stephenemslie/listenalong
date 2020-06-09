data "archive_file" "dummy" {
  type        = "zip"
  output_path = "${path.module}/lambda_function_payload.zip"

  source {
    content  = "hello"
    filename = "dummy.zip"
  }
}

resource "aws_lambda_function" "api" {
  filename      = data.archive_file.dummy.output_path
  function_name = "ListenalongAPI"
  handler       = "listenalong.lambda.handler"
  runtime       = "go1.x"
  role          = aws_iam_role.api_lambda.arn

  environment {
    variables = {
      HOST                       = "getinloser.biz"
      SOCIAL_AUTH_SPOTIFY_KEY    = "secret://listenalong-api/SOCIAL_AUTH_SPOTIFY_KEY"
      SOCIAL_AUTH_SPOTIFY_SECRET = "secret://listenalong-api/SOCIAL_AUTH_SPOTIFY_SECRET"
      API_SECRET_KEY             = "secret://listenalong-api/API_SECRET_KEY"
    }
  }

  # lifecycle {
  #   ignore_changes = [
  #     template[0].spec[0].containers[0].image,
  #     template[0].metadata[0].annotations["client.knative.dev/user-image"],
  #     template[0].metadata[0].annotations["run.googleapis.com/client-name"],
  #     template[0].metadata[0].annotations["run.googleapis.com/client-version"]
  #   ]
  # }
}
