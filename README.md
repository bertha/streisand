# StreiSANd

StreiSANd is a replicated content addressed blob store. Users can upload arbitrary data, and will get the hash of that data in response. Later they can retrieve that data by hash.

StreiSANd takes care of replicating all uploads to other StreiSANd servers, even if they currently aren't online. It tries to do so realtime, but has a smart algorithm to detect which nodes are missing which blobs which it uses to bring all nodes in sync after temporary issues.
