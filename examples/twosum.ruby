def two_sum(nums, target)
  hash = {}
  nums.each_with_index do |num, index|
    complement = target - num
    return [hash[complement], index] if hash.key?(complement)
    hash[num] = index
  end
  []
end

# Example usage:
puts two_sum([2, 7, 11, 15], 9).inspect
# Output: [0, 1]
