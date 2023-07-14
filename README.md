# Img Opt
A micro-service allowing for dynamic optimization of images. Supporting new formats such as AVIF and WebP.

> Still a WIP, may not be suited for production


## Use Case
- Host your website images and have them automatically optimized
- Integrate it with your app, removing the need to code it yourself


## Features
- Optimizations happen dynamically, requiring no extra storage
- Configurable optimization settings (using YAML)
    - Resize images
    - Convert
        - JPEG
        - WEBP
        - AVIF
    - Quality settings
- Automatic Optimization, detecting a clients supported image types
