function two_sum(nums::Vector{Int}, target::Int)
    num_to_index = Dict{Int, Int}()
    
    for (i, num) in enumerate(nums)
        complement = target - num
        if haskey(num_to_index, complement)
            return (num_to_index[complement], i)
        end
        num_to_index[num] = i
    end
    
    return nothing  # If no solution found
end
