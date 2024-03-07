---
title: an assuredly foolish venture
created: 2024-03-05T22:37:53.62147009-06:00
modified: 2024-03-06T08:13:00.611558241-06:00
id: 215c6975-fa0c-4e7f-b33a-bd22acaadb4f
style: article
feeds:
- public
checksum: sha1:5ae33ffa796b19b0889b049f38d3286d3c6c045b
---
it seems there likely isn't much value in parsing the markdown into an AST per se

despite that being the *obvious* approach.

there is certainly the possibility that just using the default parser and html renderer is good enough ... i guess

next items for attention:
* metadata-only changes don't work. we're only scanning the body for a change. we should also detect updated metadata; maybe see if the rendered yaml is byte-for-byte the same
  * it would be great if we could not rely on fs mtimes for this, but we might not have better alternatives.
  * we could determine a canonical rendering that does not include the checksum, render it, checksum it, and then include that in the updated file contents. seems very fragile though.
  * keep thinking about better options, but maybe fs mtimes are just good enough
* figure out what precisely a feed is, or if we need to abstract those behind "tags" such that a feed ingests certain tags, and then the tags define the aggregation
  * it should be possible to have ...
    * a text be explicitly not published to any feed
    * a "private" feed only avilable to readers that know the uuid of the feed
    * a "private" published text only avilable to readers that know the uuid of the text

future consideration:
* line endings should be normalized in processing
  * probably just by converting all \r\n to \n first
  * then converting all bare \r to \n second
* feeds obviously need to publish somewhere
  * atom and rss should publish to a fixed asset location and travel with rendered html
  * activitypub would be cool, but maybe awkward
    * objects can live at fixed locations: cool!
    * we have to push to other instances, but which? bleaugh!
  * email should be possible but eugh
    * this ain't for marketing
* need to provide hosting options
  * out-of-the-box we're publishing to a local folder
  * but we should offer automatic s3 (with optional cloudfront invalidation)
  * maybe r2?
  * maybe git push for github pages?
  * sftp is probably not too far to reach
  * not totally unreasonable to offer in-place file serve, but icky without https