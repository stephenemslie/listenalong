resource "aws_s3_bucket" "api" {
  bucket = "listenalong-api"
  acl    = "private"

  tags = {
    Project = "listenalong"
  }
}
