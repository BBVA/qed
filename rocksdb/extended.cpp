/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

#include "extended.h"

#include "rocksdb/c.h"
#include "_cgo_export.h"
#include "rocksdb/db.h"
#include "rocksdb/statistics.h"
#include "rocksdb/options.h"
#include "rocksdb/utilities/backupable_db.h"
#include "rocksdb/status.h"

using rocksdb::DB;
using rocksdb::ColumnFamilyHandle;
using rocksdb::Statistics;
using rocksdb::HistogramData;
using rocksdb::StatsLevel;
using rocksdb::Options;
using rocksdb::Cache;
using rocksdb::NewLRUCache;
using rocksdb::Slice;
using std::shared_ptr;
using rocksdb::BackupEngine;
using rocksdb::BackupInfo;
using rocksdb::Status;

extern "C" {

struct rocksdb_t { DB* rep; };
struct rocksdb_statistics_t { std::shared_ptr<Statistics> rep; };
struct rocksdb_histogram_data_t { rocksdb::HistogramData* rep; };
struct rocksdb_options_t { Options rep; };
struct rocksdb_cache_t { std::shared_ptr<Cache> rep; };
struct rocksdb_column_family_handle_t  { ColumnFamilyHandle* rep; };
struct rocksdb_backup_engine_t   { BackupEngine*     rep; };
struct rocksdb_backup_engine_info_t { std::vector<BackupInfo> rep; };

void rocksdb_options_set_atomic_flush(
    rocksdb_options_t* opts, unsigned char value) {
    opts->rep.atomic_flush = value;
}

rocksdb_cache_t* rocksdb_cache_create_lru_with_ratio(
    size_t capacity, double hi_pri_pool_ratio) {
    rocksdb_cache_t* c = new rocksdb_cache_t;
    c->rep = NewLRUCache(capacity, -1, false, hi_pri_pool_ratio);
    return c;
}

void rocksdb_destruct_handler(void* state) { }

rocksdb_slicetransform_t* rocksdb_slicetransform_create_ext(uintptr_t idx) {
    return rocksdb_slicetransform_create(
    	(void*)idx,
    	rocksdb_destruct_handler,
    	(char* (*)(void*, const char*, size_t, size_t*))(rocksdb_slicetransform_transform),
    	(unsigned char (*)(void*, const char*, size_t))(rocksdb_slicetransform_in_domain),
    	(unsigned char (*)(void*, const char*, size_t))(rocksdb_slicetransform_in_range),
    	(const char* (*)(void*))(rocksdb_slicetransform_name));
}


/* Backup */

static bool SaveError(char** errptr, const Status& s) {
  assert(errptr != nullptr);
  if (s.ok()) {
    return false;
  } else if (*errptr == nullptr) {
    *errptr = strdup(s.ToString().c_str());
  } else {
    free(*errptr);
    *errptr = strdup(s.ToString().c_str());
  }
  return true;
}

void rocksdb_backup_engine_create_new_backup_with_metadata(rocksdb_backup_engine_t* be,
                                             rocksdb_t* db,
                                             int num_metadata,
                                             char** app_metadata,
                                             char** errptr) {
    std::vector<std::string> meta(num_metadata);
    for (int i = 0; i < num_metadata; i++) {
        meta[i] = std::string(app_metadata[i]);
    }
    SaveError(errptr, be->rep->CreateNewBackupWithMetadata(db->rep, meta[0]));
}

static char* CopyString(const std::string& str) {
  char* result = reinterpret_cast<char*>(malloc(sizeof(char) * str.size()));
  memcpy(result, str.data(), sizeof(char) * str.size());
  return result;
}

void rocksdb_backup_engine_info_metadata(const rocksdb_backup_engine_info_t* info, 
        int index,
        char** metadata_list,
        size_t* metadata_list_size){

    std::vector<char*> cMetadata;
    for (size_t i = 0; i < info->rep[index].app_metadata.size()-1  ; i++){
        cMetadata[i] = CopyString(&info->rep[index].app_metadata[i]);
        printf("Meta %s\n", cMetadata[i]);
    }

    size_t aux = cMetadata.size();
    printf("AUX %d\n",sizeof(cMetadata));
    metadata_list_size = &aux; // &cMetadata.size(); 
    metadata_list = &cMetadata[0];
}

/* Statistics */

rocksdb_statistics_t* rocksdb_create_statistics() {
    rocksdb_statistics_t* result = new rocksdb_statistics_t;
    result->rep = rocksdb::CreateDBStatistics();
    return result;
}

int rocksdb_property_int_cf(
    rocksdb_t* db, rocksdb_column_family_handle_t* column_family,
    const char* propname, uint64_t *out_val) {
    if (db->rep->GetIntProperty(column_family->rep, Slice(propname), out_val)) {
        return 0;
    } else {
        return -1;
    }
}

void rocksdb_options_set_statistics(
    rocksdb_options_t* opts, 
    rocksdb_statistics_t* stats) {
    if (stats) {
        opts->rep.statistics = stats->rep;
    }
}

rocksdb_stats_level_t rocksdb_statistics_stats_level(
    rocksdb_statistics_t* stats) {
        return static_cast<rocksdb_stats_level_t>(stats->rep->stats_level_);
}

void rocksdb_statistics_set_stats_level(
    rocksdb_statistics_t* stats,
    rocksdb_stats_level_t level) {
        stats->rep->stats_level_ = static_cast<StatsLevel>(level);
}

void rocksdb_statistics_reset(
    rocksdb_statistics_t* stats) {
        stats->rep->Reset();
}

uint64_t rocksdb_statistics_get_ticker_count(
    rocksdb_statistics_t* stats, 
    rocksdb_tickers_t ticker_type) {
        return stats->rep->getTickerCount(ticker_type);
}

uint64_t rocksdb_statistics_get_and_reset_ticker_count(
    rocksdb_statistics_t* stats, 
    rocksdb_tickers_t ticker_type) {
        return stats->rep->getAndResetTickerCount(ticker_type);
}

void rocksdb_statistics_destroy(rocksdb_statistics_t* stats) {
    delete stats;
}

void rocksdb_statistics_histogram_data(
    const rocksdb_statistics_t* stats, 
    rocksdb_histograms_t type, 
    const rocksdb_histogram_data_t* data) {
        stats->rep->histogramData(type, data->rep);
}

// Histogram

rocksdb_histogram_data_t* rocksdb_histogram_create_data() {
    rocksdb_histogram_data_t* result = new rocksdb_histogram_data_t;
    rocksdb::HistogramData hData;
    result->rep = &hData;
    return result;
}

double rocksdb_histogram_get_average(rocksdb_histogram_data_t* data) {
    return data->rep->average;
}

double rocksdb_histogram_get_median(rocksdb_histogram_data_t* data) {
    return data->rep->median;
}

double rocksdb_histogram_get_percentile95(rocksdb_histogram_data_t* data) {
    return data->rep->percentile95;
}

double rocksdb_histogram_get_percentile99(rocksdb_histogram_data_t* data) {
    return data->rep->percentile99;
}

double rocksdb_histogram_get_stdev(rocksdb_histogram_data_t* data) {
    return data->rep->standard_deviation;
}

double rocksdb_histogram_get_max(rocksdb_histogram_data_t* data) {
    return data->rep->max;
}

uint64_t rocksdb_histogram_get_count(rocksdb_histogram_data_t* data) {
    return data->rep->count;
}

uint64_t rocksdb_histogram_get_sum(rocksdb_histogram_data_t* data) {
    return data->rep->sum;
}

void rocksdb_histogram_data_destroy(rocksdb_histogram_data_t* data);

void rocksdb_histogram_data_destroy(rocksdb_histogram_data_t* data) {
    delete data;
}

} // end extern "C"
