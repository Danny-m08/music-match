CREATE CONSTRAINT unique_email IF NOT EXISTS for (user:User) require user.email IS Unique;
CREATE CONSTRAINT unique_username IF NOT EXISTS for (user:User) require user.username IS UNIQUE;
CREATE CONSTRAINT unique_listing_ID IF NOT EXISTS for (listing:Listing) require listing.ID IS UNIQUE;