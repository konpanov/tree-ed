(* two_sum : int array -> int -> (int * int) option *)
let two_sum nums target =
  let table = Hashtbl.create (Array.length nums) in
  let rec find i =
    if i >= Array.length nums then
      None
    else
      let complement = target - nums.(i) in
      match Hashtbl.find_opt table complement with
      | Some j -> Some (j, i)
      | None ->
          Hashtbl.add table nums.(i) i;
          find (i + 1)
  in
  find 0
