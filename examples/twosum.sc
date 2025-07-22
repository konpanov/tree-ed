object TwoSum {
  def twoSum(nums: Array[Int], target: Int): Array[Int] = {
    val numMap = scala.collection.mutable.Map[Int, Int]()
    
    for (i <- nums.indices) {
      val complement = target - nums(i)
      if (numMap.contains(complement)) {
        return Array(numMap(complement), i)
      }
      numMap(nums(i)) = i
    }
    
    Array() // Return empty array if no solution found
  }

  def main(args: Array[String]): Unit = {
    val nums = Array(2, 7, 11, 15)
    val target = 9
    val result = twoSum(nums, target)
    
    println(result.mkString("[", ", ", "]"))  // Output: [0, 1]
  }
}
