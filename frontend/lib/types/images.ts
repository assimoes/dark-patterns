// POST /api/images
export type UploadImageItem = {
    file: File;
    description?: string;
};

export type UploadImagesInput = {
    game_id: number;
    images: UploadImageItem[];
};

export type UploadImagesResult = {
    uploaded: number;
    game_id: number;
    image_ids: number[];
};