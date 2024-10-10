terraform {
  backend "http" {
    username       = "tf-urmzkdw5z18cbu"
    password       = "9SU$XuCUrLvy8BsW$UjgHMN0"
    address        = "http://localhost:8000/client/c9vxgddyr7c/state"
    lock_address   = "http://localhost:8000/client/c9vxgddyr7c/lock"
    unlock_address = "http://localhost:8000/client/c9vxgddyr7c/unlock"
    lock_method    = "POST"
    unlock_method  = "POST"
  }
}