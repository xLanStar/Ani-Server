# 資料格式

## user
- String Account
- String Password
- String UserName
- String Introduction
- uint32 LikeMedias[]
- uint32 LikeReviews[]

## review
- uint32 Id
- uint8 Rank
- String Content

## media
- uint8 Type
    ### 類型 - 0[動畫]
    - uint16 Episodes(總集數) 從 Episodes>>15 開始 
        - byte Videos[80 * Episodes]
        - uint32 ExEpisodes[]
    - uint32 Id_if101
    ### 類型 - 1[漫畫]
    - uint32 volumes[]
        [      起點       ] [       終點       ]
        0000 0000 0000 0000 0000 0000 0000 0000
        0000 0000 0000 0001 0000 0000 0001 0000 -> 有 1  ~  16
        0000 0000 0011 0000 0000 0000 0100 0000 -> 有 48 ~  64
    - uint32 Id_cartoonmad
    ### 類型 - 1[小說]
    - uint16 volumes
- string Title_tw
- string Description