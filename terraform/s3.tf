resource "aws_s3_bucket" "listenalong_db" {
  bucket = "listenalong-db"
  acl    = "private"

  tags = {
    Project = "listenalong"
  }
}
