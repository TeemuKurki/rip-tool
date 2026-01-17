/*
 * Portions of this code are adapted from libbluray:
 *   tools/db_list_titles.c
 * Copyright (C) 2009-2010 John Stebbins
 *
 * libbluray is licensed under the GNU Lesser General Public License (LGPL) v2.1 or later.
 * This file is distributed under the same license.
 * 
 * Modification:
 *  - Removed input arguments
 *  - Removed title information printing
 *  - Converted code to generate JSON from chapter informations
 */


 
#include "libbluray/bluray.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>


/**
 * Create JSON Array string of each chapter information
 */
char* chapters_json(const BLURAY_TITLE_INFO *ti, int index) {
    int chapter_count = (int) ti->chapter_count;
    
    // Calculate chapter block memory size dynamically
    size_t buf_size = 2; // for start '[' and end ']'
    for (int j = 0; j < chapter_count; j++) {
        // Return required memory size and ignore actual string
        buf_size += snprintf(NULL, 0, "{\"id\":%d,\"duration\":%d,\"title\":%d},", j, (int)ti->chapters[j].duration, index);
    }

    char *json = malloc(buf_size);
    if (!json) return NULL;

    size_t off = 0;
    off += snprintf(json + off, buf_size - off, "[");
    for (int i = 0; i < chapter_count; i++) {
        off += snprintf(
            json + off, buf_size - off,
            "%s{\"id\":%d,\"duration\":%d,\"title\":%d}",
            i > 0 ? "," : "",
            i,
            (int) ti->chapters[i].duration,
            index
        );
    }

    off += snprintf(json + off, buf_size - off, "]");
    return json; // <- caller must free
}

/**
 * Join list of chapter information into single string.
 * Returning single string instead of array is easier to handle and mashall on Golang side.
 */
char *join_chapters(char **chapter_array, int count) {
    // If count of chapters is 0 return null
    if (count == 0) {
        return NULL;
    }

    // Compute total length of string
    size_t total_len = 2; // for '[' and ']'
    for (int i = 0; i < count; i++) {
        if(chapter_array[i] != NULL){
            total_len += strlen(chapter_array[i]);
            if (i < count - 1) total_len++; // commas
        }
    }
    total_len += 1; // null terminator

    // Allocate buffer for joined chapters string
    char *json_all = malloc(total_len);
    if (!json_all) {
        return NULL;
    }

    size_t off = 0;
    // Safely add '[' to json string
    off += snprintf(json_all + off, total_len - off, "[");
    for (int i = 0; i < count; i++) {
        if (i > 0) off += snprintf(json_all + off, total_len - off, ",");
        // Safely add chapter information to json string
        off += snprintf(json_all + off, total_len - off, "%s", chapter_array[i]);
    }
    // Safely add ']' to end of json string 
    snprintf(json_all + off, total_len - off, "]");

    return json_all; // caller must free
}

char *chapters(char *path) {
    BLURAY *bd;
    int count, i;
    //TODO: Pass option to filter by length 
    unsigned int seconds = 0;
    char *bd_dir = NULL;

    bd_dir = path;

    bd = bd_open(bd_dir, NULL);
    if (!bd) {
        fprintf(stderr, "bd_open(%s) failed\n", bd_dir);
        return NULL;
    }
    count = bd_get_titles(bd, TITLES_RELEVANT, seconds);
    char **chapter_array = malloc(count * sizeof(char*));
    if (!chapter_array) { 
        free(chapter_array); 
        bd_close(bd);
        return NULL;
    }

    for (i = 0; i < count; i++) {
        BLURAY_TITLE_INFO* ti;
        ti = bd_get_title_info(bd, i, 0);
        
        char *chapter_info = chapters_json(ti, i + 1);
        
        // Free title info memory
        bd_free_title_info(ti);

        size_t len = strlen(chapter_info);
        chapter_array[i] = malloc(len + 1);
        if (!chapter_array[i]) { 
            free(chapter_info); 
            continue;
            //TODO: handle error
        }
        memcpy(chapter_array[i], chapter_info, len + 1);
        // Free chapter information json array memory
        free(chapter_info);
        
    }

    // Clean up Bluray disk read
    bd_close(bd);

    char *response = join_chapters(chapter_array, count);

    // Free chapter array and items inside chapter array
    for (int j = 0; j < count; j++){
        free(chapter_array[j]);
    }
    free(chapter_array);
    
    return response;
}

char *run_chapters(const char *path) {
    return chapters((char *)path);
}
