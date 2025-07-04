#!/bin/bash

echo "Generating 40,000 dummy files..."
for i in $(seq 0 39); do # 40 subdirectories
    mkdir test_source/dir_$i
    for j in $(seq 1 1000); do # 1000 files per subdirectory
        echo "This is file $j in dir $i" > test_source/dir_$i/file_$i_dir_$j.txt
        if (( j % 5 == 0 )); then # Add some images
            echo "This is image $j" > test_source/dir_$i/image_$i_dir_$j.jpg
        fi
        if (( j % 10 == 0 )); then # Add some videos
            echo "This is video $j" > test_source/dir_$i/video_$i_dir_$j.mp4
        fi
    done
done
echo "File generation complete."
