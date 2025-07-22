twoSum :: [Int] -> Int -> [(Int, Int)]
twoSum nums target = [ (x, y) | (i, x) <- indexedNums,
                                (j, y) <- indexedNums,
                                i < j,
                                x + y == target ]
  where
    indexedNums = zip [0..] nums
