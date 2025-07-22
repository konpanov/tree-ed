#include <stdio.h>
#include <stdlib.h>

typedef struct {
    int key;
    int value;
} HashItem;

typedef struct {
    HashItem* items;
    int size;
} HashTable;

// Simple hash function
int hash(int key, int size) {
    return abs(key) % size;
}

// Create hash table
HashTable* createTable(int size) {
    HashTable* table = (HashTable*)malloc(sizeof(HashTable));
    table->size = size;
    table->items = (HashItem*)calloc(size, sizeof(HashItem));
    return table;
}

// Insert into hash table
void insert(HashTable* table, int key, int value) {
    int index = hash(key, table->size);
    while (table->items[index].value != 0) {
        index = (index + 1) % table->size; // Linear probing
    }
    table->items[index].key = key;
    table->items[index].value = value;
}

// Search in hash table
int search(HashTable* table, int key) {
    int index = hash(key, table->size);
    while (table->items[index].value != 0) {
        if (table->items[index].key == key) {
            return table->items[index].value;
        }
        index = (index + 1) % table->size;
    }
    return -1;
}

// Two Sum function
int* twoSum(int* nums, int numsSize, int target, int* returnSize) {
    HashTable* table = createTable(numsSize * 2);
    int* result = (int*)malloc(2 * sizeof(int));

    for (int i = 0; i < numsSize; i++) {
        int complement = target - nums[i];
        int index = search(table, complement);

        if (index != -1) {
            result[0] = index;
            result[1] = i;
            *returnSize = 2;
            free(table->items);
            free(table);
            return result;
        }
        insert(table, nums[i], i);
    }

    *returnSize = 0;
    free(table->items);
    free(table);
    return NULL;
}
