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
#include "rocksdb/write_batch.h"

using rocksdb::DB;
using rocksdb::ColumnFamilyHandle;
using rocksdb::Statistics;
using rocksdb::HistogramData;
using rocksdb::StatsLevel;
using rocksdb::Options;
using rocksdb::Cache;
using rocksdb::NewLRUCache;
using rocksdb::Slice;
using rocksdb::WriteBatch;
using std::shared_ptr;
using rocksdb::BackupEngine;
using rocksdb::BackupInfo;
using rocksdb::Status;
using rocksdb::RestoreOptions;

extern "C" {

struct rocksdb_t { DB* rep; };
struct rocksdb_statistics_t { std::shared_ptr<Statistics> rep; };
struct rocksdb_histogram_data_t { rocksdb::HistogramData* rep; };
struct rocksdb_options_t { Options rep; };
struct rocksdb_cache_t { std::shared_ptr<Cache> rep; };
struct rocksdb_column_family_handle_t  { ColumnFamilyHandle* rep; };
struct rocksdb_backup_engine_t   { BackupEngine*     rep; };
struct rocksdb_backup_engine_info_t { std::vector<BackupInfo> rep; };
struct rocksdb_restore_options_t { RestoreOptions rep; };
struct rocksdb_writebatch_t { WriteBatch rep; };

struct rocksdb_writebatch_handler_t : public WriteBatch::Handler {
    void* state_;
    void (*destructor_)(void*);
    void (*log_data_)(void*, const char* blob, size_t length);

    ~rocksdb_writebatch_handler_t() override { (*destructor_)(state_); }

    void LogData(const Slice& blob) override {
        (*log_data_)(state_, blob.data(), blob.size());
    }

};

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

rocksdb_writebatch_handler_t* rocksdb_writebatch_handler_create(
    void* state,
    void (*destructor)(void*),
    void (*log_data)(void*, const char* blob, size_t length)) {

    rocksdb_writebatch_handler_t* result = new rocksdb_writebatch_handler_t;
    result->state_ = state;
    result->destructor_ = destructor;
    result->log_data_ = log_data;
    return result;
}

rocksdb_writebatch_handler_t* rocksdb_writebatch_handler_create_ext(uintptr_t idx) {
    return rocksdb_writebatch_handler_create(
        (void*)idx,
        rocksdb_destruct_handler,
        (void (*)(void*, const char*, size_t))(rocksdb_writebatch_handler_log_data));
}

void rocksdb_writebatch_handler_destroy(rocksdb_writebatch_handler_t* handler) {
    delete handler;
}

void rocksdb_writebatch_iterate_ext(
    rocksdb_writebatch_t* b, 
    rocksdb_writebatch_handler_t* h) {
    b->rep.Iterate(h);
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
                                             char* app_metadata,
                                             char** errptr) {
    SaveError(errptr, be->rep->CreateNewBackupWithMetadata(db->rep, std::string(app_metadata)));
}

char* rocksdb_backup_engine_info_metadata(const rocksdb_backup_engine_info_t* info, 
        int index){
    return strdup(info->rep[index].app_metadata.c_str());
}

void rocksdb_backup_engine_restore_db_from_backup(
    rocksdb_backup_engine_t* be, uint32_t backupID, const char* db_dir, const char* wal_dir,
    const rocksdb_restore_options_t* restore_options, char** errptr) {
  SaveError(errptr, be->rep->RestoreDBFromBackup(backupID,
                                                 std::string(db_dir),
                                                 std::string(wal_dir),
                                                 restore_options->rep));
}

extern void rocksdb_backup_engine_delete_backup(
    rocksdb_backup_engine_t* be, uint32_t backupID, char** errptr){
  SaveError(errptr, be->rep->DeleteBackup(backupID));
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
