numbers=(1 2 3 4 5 6 7 8 9 10)

for num in "${numbers[@]}"; do
    if (( num % 2 == 0 )); then
        echo "$num"
    fi
done
