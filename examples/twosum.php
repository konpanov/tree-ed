<?php
function twoSum($nums, $target) {
    $map = [];  // associative array to store number => index

    foreach ($nums as $index => $num) {
        $complement = $target - $num;
        if (isset($map[$complement])) {
            return [$map[$complement], $index];
        }
        $map[$num] = $index;
    }

    return [];  // return empty array if no solution
}

// Example usage:
$numbers = [2, 7, 11, 15];
$target = 9;
$result = twoSum($numbers, $target);

print_r($result);  // Output: [0, 1]
?>
